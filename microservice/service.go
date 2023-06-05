package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	status "google.golang.org/grpc/status"
)

type ACL map[string][]string

type AdminManager struct {
	ctx    context.Context
	mu     *sync.Mutex
	acl    ACL
	events []chan *Event

	UnimplementedAdminServer
}

type BizManager struct {
	ctx context.Context
	am  *AdminManager

	UnimplementedBizServer
}

// ========================================AdminManager========================================
func NewAdminManager(ctx context.Context, aclData string) (*AdminManager, error) {
	var acls ACL
	if errUnmrshl := json.Unmarshal([]byte(aclData), &acls); errUnmrshl != nil {
		return nil, errUnmrshl
	}
	adminManager := &AdminManager{ctx: ctx, mu: new(sync.Mutex), acl: acls}

	return adminManager, nil
}

func (am *AdminManager) ReqAuth(ctx context.Context, fullMethod string) error {
	md, okMeta := metadata.FromIncomingContext(ctx)
	if !okMeta {
		return status.Errorf(codes.InvalidArgument, "Metadata is failed")
	}
	consumer, okConsumer := md["consumer"]
	if !okConsumer || len(consumer) == 0 {
		return status.Errorf(codes.Unauthenticated, "Consumer not found")
	}

	allowMethods, okACL := am.acl[consumer[0]]
	if !okACL {
		return status.Errorf(codes.Unauthenticated, "Unknown consumer")
	}
	allowed := false
	for _, method := range allowMethods {
		if ((method[len(method)-1] == '*') && strings.HasPrefix(fullMethod, method[:len(method)-1])) ||
			(method == fullMethod) {
			allowed = true
			break
		}
	}
	if !allowed {
		return status.Errorf(codes.Unauthenticated, "Method not allowed")
	}

	client, errClient := peer.FromContext(ctx)
	if !errClient {
		return status.Errorf(codes.Unauthenticated, "Bad context")
	}
	newEvent := new(Event)
	newEvent.Timestamp = time.Now().Unix()
	newEvent.Consumer = consumer[0]
	newEvent.Method = fullMethod
	newEvent.Host = client.Addr.String()

	am.mu.Lock()
	defer am.mu.Unlock()
	for _, event := range am.events {
		if event != nil {
			time.Sleep(10 * time.Microsecond)
			event <- newEvent
		}
	}

	return nil
}

func (am *AdminManager) Logging(in *Nothing, out Admin_LoggingServer) error {
	ch, idx := am.CreateChannel()
	defer am.CloseChannel(idx)
	for {
		select {
		case event := <-ch:
			err := out.Send(event)
			if err != nil {
				return err
			}

		case <-am.ctx.Done():
			return nil
		}
	}
}

func (am *AdminManager) Statistics(in *StatInterval, out Admin_StatisticsServer) error {
	ch, chIdx := am.CreateChannel()
	ticker := time.NewTicker(time.Duration(in.IntervalSeconds) * time.Second)
	defer func() {
		am.CloseChannel(chIdx)
		ticker.Stop()
	}()
	stat := new(Stat)
	stat.Timestamp = time.Now().Unix()
	stat.ByMethod = map[string]uint64{}
	stat.ByConsumer = map[string]uint64{}
	for {
		select {
		case event := <-ch:
			stat.ByConsumer[event.Consumer]++
			stat.ByMethod[event.Method]++

		case <-ticker.C:
			err := out.Send(stat)
			if err != nil {
				return err
			}
			stat.Timestamp = time.Now().Unix()
			stat.ByMethod = map[string]uint64{}
			stat.ByConsumer = map[string]uint64{}

		case <-am.ctx.Done():
			return nil
		}
	}
}

func (am *AdminManager) CreateChannel() (chan *Event, int) {
	am.mu.Lock()
	defer am.mu.Unlock()
	for idx, event := range am.events {
		if event == nil {
			am.events[idx] = make(chan *Event)
			return am.events[idx], idx
		}
	}
	ch := make(chan *Event)
	am.events = append(am.events, ch)
	return ch, len(am.events) - 1
}

func (am *AdminManager) CloseChannel(idx int) {
	am.mu.Lock()
	defer am.mu.Unlock()
	close(am.events[idx])
	am.events[idx] = nil
}

// =========================================BizManager=========================================
func (bm *BizManager) Check(ctx context.Context, in *Nothing) (*Nothing, error) {
	return in, nil
}

func (bm *BizManager) Add(ctx context.Context, in *Nothing) (*Nothing, error) {
	return in, nil
}

func (bm *BizManager) Test(ctx context.Context, in *Nothing) (*Nothing, error) {
	return in, nil
}

// ========================================Microservice========================================
func StartMyMicroservice(ctx context.Context, listenAddr, aclData string) error {

	adminManager, errAdminCreate := NewAdminManager(ctx, aclData)
	if errAdminCreate != nil {
		return errAdminCreate
	}
	bizManager := &BizManager{ctx: ctx, am: adminManager}

	unaryInterceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		err := adminManager.ReqAuth(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}

	streamInterceptor := func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		err := adminManager.ReqAuth(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)
	RegisterBizServer(server, bizManager)
	RegisterAdminServer(server, adminManager)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Panicln("can't listen port", listenAddr, err)
	}

	go func(ctx context.Context, server *grpc.Server) {
		for range ctx.Done() {
		}
		server.GracefulStop()
	}(ctx, server)

	go func(server *grpc.Server) {
		if err := server.Serve(lis); err != nil {
			log.Panicln("Server stop", err)
		}
	}(server)

	return nil
}
