package main

// тут писать код тестов
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestNegativeLimit(t *testing.T) {
	expectedError := fmt.Errorf("limit must be > 0")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      -1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestNegativeOffset(t *testing.T) {
	expectedError := fmt.Errorf("offset must be > 0")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     -1,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBoundLimit(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      30,
		Offset:     1,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if len(result.Users) != 25 {
		t.Errorf("Wrong number of users, expected 25, got %#v", len(result.Users))
	}
}

func TestLimitXmlRecords(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      25,
		Offset:     25,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if len(result.Users) != 10 {
		t.Errorf("Wrong number of users, expected 35, got %#v", len(result.Users))
	}
}

func TestBadToken(t *testing.T) {
	expectedError := fmt.Errorf("bad AccessToken")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenBad",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestTimeOut(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	expectedError := fmt.Errorf("timeout for limit=2&offset=0&order_by=1&order_field=ID&query=")
	_, err := client.FindUsers(req)

	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBadUrl(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         "BadURL",
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	expectedError := fmt.Errorf("unknown error Get \"BadURL?limit=2&offset=0&order_by=1&order_field=ID&query=\": unsupported protocol scheme \"\"")
	_, err := client.FindUsers(req)

	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBadJsonUnpack(t *testing.T) {

	badReq := "Bad_information"
	badReqType := reflect.TypeOf(badReq)
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			result, errM := json.Marshal(badReq)
			if errM != nil {
				return
			}
			_, err := w.Write(result)
			if err != nil {
				return
			}
		}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	expectedError := fmt.Errorf("cant unpack error json: json: cannot unmarshal %s into Go value of type main.SearchErrorResponse", badReqType)
	_, err := client.FindUsers(req)

	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestUnknownBadRequest(t *testing.T) {

	badReq := "Bad request 666"
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			resp := SearchErrorResponse{}
			resp.Error = badReq
			result, errM := json.Marshal(resp)
			if errM != nil {
				return
			}
			_, err := w.Write(result)
			if err != nil {
				return
			}
		}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	expectedError := fmt.Errorf("unknown bad request error: %s", badReq)
	_, err := client.FindUsers(req)

	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBadUnpackResultJson(t *testing.T) {

	badReq := "Bad_information"
	badReqType := reflect.TypeOf(badReq)
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			result, errM := json.Marshal(badReq)
			if errM != nil {
				return
			}
			_, err := w.Write(result)
			if err != nil {
				return
			}
		}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	expectedError := fmt.Errorf("cant unpack result json: json: cannot unmarshal %s into Go value of type []main.User", badReqType)
	_, err := client.FindUsers(req)

	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBadOrderField(t *testing.T) {
	expectedError := fmt.Errorf("OrderFeld %s invalid", "incorrect")
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "incorrect",
		OrderBy:    1,
	}
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestCorrectWorkIDAsc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     0,
				Name:   "BoydWolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male",
			},
			{
				ID:     1,
				Name:   "HildaMayer",
				Age:    21,
				About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
				Gender: "female",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkIDDesc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     34,
				Name:   "KaneSharp",
				Age:    34,
				About:  "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n",
				Gender: "male",
			},
			{
				ID:     33,
				Name:   "TwilaSnow",
				Age:    36,
				About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
				Gender: "female",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    -1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkNameAsc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     16,
				Name:   "AnnieOsborn",
				Age:    35,
				About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
				Gender: "female",
			},
			{
				ID:     19,
				Name:   "BellBauer",
				Age:    26,
				About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     1,
		Query:      "",
		OrderField: "Name",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkNameDesc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     33,
				Name:   "TwilaSnow",
				Age:    36,
				About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
				Gender: "female",
			},
			{
				ID:     18,
				Name:   "TerrellHall",
				Age:    27,
				About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     1,
		Query:      "",
		OrderField: "Name",
		OrderBy:    -1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkAgeAsc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     1,
				Name:   "HildaMayer",
				Age:    21,
				About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
				Gender: "female",
			},
			{
				ID:     15,
				Name:   "AllisonValdez",
				Age:    21,
				About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkAgeDesc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     32,
				Name:   "ChristyKnapp",
				Age:    40,
				About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
				Gender: "female",
			},
			{
				ID:     13,
				Name:   "WhitleyDavidson",
				Age:    40,
				About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    -1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkEmptyAsc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     16,
				Name:   "AnnieOsborn",
				Age:    35,
				About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
				Gender: "female",
			},
			{
				ID:     19,
				Name:   "BellBauer",
				Age:    26,
				About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkEmptyDesc(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     33,
				Name:   "TwilaSnow",
				Age:    36,
				About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
				Gender: "female",
			},
			{
				ID:     18,
				Name:   "TerrellHall",
				Age:    27,
				About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     1,
		Query:      "",
		OrderField: "",
		OrderBy:    -1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectWorkIDAscQueryExist(t *testing.T) {
	Response := &SearchResponse{
		Users: []User{
			{
				ID:     0,
				Name:   "BoydWolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male",
			},
			{
				ID:     3,
				Name:   "EverettDillard",
				Age:    27,
				About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
				Gender: "male",
			},
		},
		NextPage: true,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "esse",
		OrderField: "ID",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestCorrectEmptyOutput(t *testing.T) {
	Response := &SearchResponse{
		Users:    []User(nil),
		NextPage: false,
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "Unexistwordhere",
		OrderField: "ID",
		OrderBy:    1,
	}
	result, err := client.FindUsers(req)

	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(Response, result) {
		t.Errorf("Wrong result, expected %#v, got %#v", Response, result)
	}
}

func TestAtoiCheckLimit(t *testing.T) {
	searcherParams := url.Values{}
	searcherParams.Add("limit", "fake")
	searcherParams.Add("offset", "1")
	searcherParams.Add("query", "")
	searcherParams.Add("order_field", "ID")
	searcherParams.Add("order_by", "1")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searcherReq, _ := http.NewRequest("GET", ts.URL+"?"+searcherParams.Encode(), nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", "AccessTokenGood")
	client = &http.Client{Timeout: time.Second}

	resp, err := client.Do(searcherReq)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode: %#v", resp.StatusCode)
	}
}

func TestAtoiCheckOffset(t *testing.T) {
	searcherParams := url.Values{}
	searcherParams.Add("limit", "1")
	searcherParams.Add("offset", "fake")
	searcherParams.Add("query", "")
	searcherParams.Add("order_field", "ID")
	searcherParams.Add("order_by", "1")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searcherReq, _ := http.NewRequest("GET", ts.URL+"?"+searcherParams.Encode(), nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", "AccessTokenGood")
	client = &http.Client{Timeout: time.Second}

	resp, err := client.Do(searcherReq)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode: %#v", resp.StatusCode)
	}
}

func TestAtoiCheckOrderBy(t *testing.T) {
	searcherParams := url.Values{}
	searcherParams.Add("limit", "2")
	searcherParams.Add("offset", "1")
	searcherParams.Add("query", "")
	searcherParams.Add("order_field", "ID")
	searcherParams.Add("order_by", "fake")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searcherReq, _ := http.NewRequest("GET", ts.URL+"?"+searcherParams.Encode(), nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", "AccessTokenGood")
	client = &http.Client{Timeout: time.Second}

	resp, err := client.Do(searcherReq)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode: %#v", resp.StatusCode)
	}
}

func TestAtoiNegativeLimit(t *testing.T) {
	searcherParams := url.Values{}
	searcherParams.Add("limit", "-1")
	searcherParams.Add("offset", "1")
	searcherParams.Add("query", "")
	searcherParams.Add("order_field", "ID")
	searcherParams.Add("order_by", "1")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searcherReq, _ := http.NewRequest("GET", ts.URL+"?"+searcherParams.Encode(), nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", "AccessTokenGood")
	client = &http.Client{Timeout: time.Second}

	resp, err := client.Do(searcherReq)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode: %#v", resp.StatusCode)
	}
}

func TestAtoiNegativeOffset(t *testing.T) {
	searcherParams := url.Values{}
	searcherParams.Add("limit", "1")
	searcherParams.Add("offset", "-1")
	searcherParams.Add("query", "")
	searcherParams.Add("order_field", "ID")
	searcherParams.Add("order_by", "1")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	searcherReq, _ := http.NewRequest("GET", ts.URL+"?"+searcherParams.Encode(), nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", "AccessTokenGood")
	client = &http.Client{Timeout: time.Second}

	resp, err := client.Do(searcherReq)
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode: %#v", resp.StatusCode)
	}
}

func TestFakeXmlRead(t *testing.T) {
	expectedError := fmt.Errorf("SearchServer fatal error")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    0,
	}
	dataFileName = "fake_dataset.xml"
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}

func TestBadXmlRead(t *testing.T) {
	expectedError := fmt.Errorf("SearchServer fatal error")

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "AccessTokenGood",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      1,
		Offset:     0,
		Query:      "",
		OrderField: "ID",
		OrderBy:    0,
	}
	dataFileName = "broke_dataset.xml"
	_, err := client.FindUsers(req)
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error: %#v", err)
	}
}
