package main

// тут писать SearchServer
import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

var dataFileName = "dataset.xml"

type Row struct {
	XMLName   xml.Name `xml:"row"`
	ID        int      `xml:"id"`
	Age       int      `xml:"age"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	Gender    string   `xml:"gender"`
	About     string   `xml:"about"`
}

type Root struct {
	XMLName xml.Name `xml:"root"`
	Rows    []Row    `xml:"row"`
}

func BadStatusResp(w http.ResponseWriter) {
	emptyUser, errJSON := json.Marshal([]User{})
	if errJSON != nil {
		log.Printf("Bad pack json. Err:%v", errJSON)
	}
	_, errWrite := w.Write(emptyUser)
	if errWrite != nil {
		log.Printf("Bad write response. Err:%v", errWrite)
	}
}

func CaseSort(orderField string, orderBy int, users []User) bool {
	// Сортируем по определенному полю в порядке OrderBy
	switch {
	case orderField == "ID":
		if orderBy == 1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].ID < users[j].ID
			})
		}
		if orderBy == -1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].ID > users[j].ID
			})
		}

	case orderField == "Name" || orderField == "":
		// Также при пустом аналогично
		if orderBy == 1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name < users[j].Name
			})
		}
		if orderBy == -1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].Name > users[j].Name
			})
		}

	case orderField == "Age":
		if orderBy == 1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age < users[j].Age
			})
		}
		if orderBy == -1 {
			sort.Slice(users, func(i, j int) bool {
				return users[i].Age > users[j].Age
			})
		}

	default:
		return true
	}
	return false
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	accessToken := r.Header.Get("AccessToken")

	if accessToken != "AccessTokenGood" {
		log.Printf("401 Unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		BadStatusResp(w)
		return
	}

	limit, errLimit := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, errOffset := strconv.Atoi(r.URL.Query().Get("offset"))
	orderBy, errOrderBy := strconv.Atoi(r.URL.Query().Get("order_by"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")

	if errLimit != nil {
		log.Printf("400 Bad Request. Err:%v", errLimit)
		w.WriteHeader(http.StatusBadRequest)
		BadStatusResp(w)
		return
	}
	if errOffset != nil {
		log.Printf("400 Bad Request. Err:%v", errOffset)
		w.WriteHeader(http.StatusBadRequest)
		BadStatusResp(w)
		return
	}
	if errOrderBy != nil {
		log.Printf("400 Bad Request. Err:%v", errOffset)
		w.WriteHeader(http.StatusBadRequest)
		BadStatusResp(w)
		return
	}
	if limit < 0 {
		log.Printf("400 Bad Request. Negative limit")
		w.WriteHeader(http.StatusBadRequest)
		BadStatusResp(w)
		return
	}
	if offset < 0 {
		log.Printf("400 Bad Request. Negative offset")
		w.WriteHeader(http.StatusBadRequest)
		BadStatusResp(w)
		return
	}

	dataXML, errFile := ioutil.ReadFile(dataFileName)
	if errFile != nil {
		log.Printf("500 Internal Server Error. Err:%v", errOffset)
		w.WriteHeader(http.StatusInternalServerError)
		BadStatusResp(w)
		return
	}
	data := new(Root)
	err := xml.Unmarshal(dataXML, &data)
	if err != nil {
		log.Printf("500 Internal Server Error. Err:%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		BadStatusResp(w)
		return
	}

	var users []User
	// Ищем подстроку query в записях и записываем полученную инфу в юзера
	for _, record := range data.Rows {
		if strings.Contains(record.About, query) ||
			strings.Contains(record.FirstName+record.LastName, query) {
			users = append(users,
				User{
					ID:     record.ID,
					Name:   record.FirstName + record.LastName,
					Age:    record.Age,
					About:  record.About,
					Gender: record.Gender,
				},
			)
		}
	}

	// Сортируем по определенному полю в порядке OrderBy
	if CaseSort(orderField, orderBy, users) {
		w.WriteHeader(http.StatusBadRequest)
		resp := SearchErrorResponse{}
		resp.Error = ErrorBadOrderField
		result, errJSON := json.Marshal(resp)
		if errJSON != nil {
			log.Printf("Bad pack json. Err:%v", errJSON)
		}
		_, err := w.Write(result)
		if err != nil {
			log.Printf("err:%v", err)
		}
		return
	}

	// Получаем финальный срез
	if (offset + limit) > len(users) {
		users = users[offset:]
	} else {
		users = users[offset : offset+limit]
	}

	w.WriteHeader(http.StatusOK)
	result, errJSON := json.Marshal(users)
	_, errWrite := w.Write(result)
	if errJSON != nil {
		log.Printf("Bad pack json. Err:%v", errJSON)
	}
	if errWrite != nil {
		log.Printf("Bad write response. Err:%v", errWrite)
	}
}
