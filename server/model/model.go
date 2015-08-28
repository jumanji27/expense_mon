package model

import (
    "fmt"
    "time"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "strings"
    // "reflect"

    "gopkg.in/mgo.v2"

    "money_mon/server/helpers"
)


type Main struct {
    Helpers helpers.Main
    MongoCollection *mgo.Collection
    MongoSession *mgo.Session
    Expense
}

const (
    LogTimeFormat = "02 Jan 2006 15:04:05"
)


func (self *Main) Init() {
    self.Helpers = helpers.Main{}

    session, err := mgo.Dial("localhost:27017")
    if err != nil {
        self.Helpers.LogError(err)
    }

    self.MongoSession = session
    self.MongoCollection = session.DB("test").C("money_mon")

    self.Helpers.LogSimpleMessage("Mongo ready")
}

type Expense struct {
    Date time.Time
    ExpensePlain
}

type ExpensePlain struct {
    Value int `json:"value"`
    Comment string `json:"comment"`
}

func (self *Main) Get() map[string]interface{} {
    db_expenses := []Expense{}
    self.MongoCollection.Find(nil).All(&db_expenses)

    api_expenses_month := []map[string]interface{}{}
    api_expenses_year := [][]map[string]interface{}{}
    api_expenses := [][][]map[string]interface{}{}

    current_loop_month := db_expenses[0].Date.Month()
    current_loop_year := db_expenses[0].Date.Year()

    for db_expense_itr := 0; db_expense_itr < len(db_expenses); db_expense_itr++ {
        if db_expenses[db_expense_itr].Date.Month() != current_loop_month {
            api_expenses_year = append(api_expenses_year, api_expenses_month)

            if db_expenses[db_expense_itr].Date.Year() != current_loop_year {
                api_expenses = append(api_expenses, api_expenses_year)

                api_expenses_year = [][]map[string]interface{}{}
            }

            api_expenses_month = []map[string]interface{}{}
        }

        week_number := db_expenses[db_expense_itr].Date.Day() / 7

        if week_number == 0 {
            week_number = 1
        }

        api_expense := map[string]interface{}{}

        if len(db_expenses[db_expense_itr].Comment) > 0 {
            api_expense = map[string]interface{}{
                "week": week_number,
                "value": db_expenses[db_expense_itr].Value,
                "comment": db_expenses[db_expense_itr].Comment,
            }
        } else {
            api_expense = map[string]interface{}{
                "week": week_number,
                "value": db_expenses[db_expense_itr].Value,
            }
        }

        api_expenses_month = append(api_expenses_month, api_expense)

        current_loop_month = db_expenses[db_expense_itr].Date.Month()
        current_loop_year = db_expenses[db_expense_itr].Date.Year()

        // Last iteration
        if db_expense_itr + 1 == len(db_expenses) {
            api_expenses = append(api_expenses, api_expenses_year)
        }
    }

    return map[string]interface{}{
        "success": api_expenses,
        "error": nil,
    }
}

func (self *Main) Set(res *http.Request) map[string]interface{} {
    body_uint8, err := ioutil.ReadAll(res.Body)
    if err != nil {
        self.Helpers.LogError(err)
    }

    body := strings.Replace(string(body_uint8), "'", "\"", -1)

    db_expense_plain := ExpensePlain{}

    err = json.Unmarshal([]byte(body), &db_expense_plain)
    if err != nil {
        self.Helpers.LogError(err)
    }

    self.MongoCollection.Insert(
        &Expense{time.Now(), db_expense_plain},
    )

    // No Generics T_T
    fmt.Printf(
        "%s | Added to DB: %s\n",
        time.Now().Format(LogTimeFormat),
        db_expense_plain,
    )

    return map[string]interface{}{
        "success": true,
        "error": nil,
    }
}