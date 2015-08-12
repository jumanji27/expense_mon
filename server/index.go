package main


import (
    "fmt"
    "log"
    "time"
    // "reflect"

    "encoding/json"

    "github.com/go-martini/martini"
    "github.com/martini-contrib/render"

    "gopkg.in/mgo.v2"
    // "gopkg.in/mgo.v2/bson"
)


type Index struct {
    MongoCollection *mgo.Collection
    MongoSession *mgo.Session
    DBExpense
    APIWeek
}


func (self *Index) db_init() {
    session, err := mgo.Dial("localhost:27017")
    if err != nil {
        log.Fatal(err)
    }

    self.MongoSession = session
    self.MongoCollection = session.DB("test").C("money_mon")
}

type DBExpense struct {
    Date time.Time
    Value int
    Comment string
}

type APIWeek struct {
    Week int
    Value int
    Comment string
}

func (self *Index) db_get() string {
    defer self.MongoSession.Close()

    db_expenses := []DBExpense{}
    self.MongoCollection.Find(nil).All(&db_expenses)

    api_expenses_month := []APIWeek{}
    api_expenses_year := [][]APIWeek{}
    api_expenses := [][][]APIWeek{}

    current_loop_month := db_expenses[0].Date.Month()
    current_loop_year := db_expenses[0].Date.Year()

    for db_expense_itr := 0; db_expense_itr < len(db_expenses); db_expense_itr++ {
        if db_expenses[db_expense_itr].Date.Month() != current_loop_month {
            api_expenses_year = append(api_expenses_year, api_expenses_month)

            if db_expenses[db_expense_itr].Date.Year() != current_loop_year {
                api_expenses = append(api_expenses, api_expenses_year)

                api_expenses_year = [][]APIWeek{}
            }

            api_expenses_month = []APIWeek{}
        }

        api_expenses_month = append(
            api_expenses_month,
            APIWeek{1, db_expenses[db_expense_itr].Value,db_expenses[db_expense_itr].Comment}, // TODO: 1 — hardcode
        )

        current_loop_month = db_expenses[db_expense_itr].Date.Month()
        current_loop_year = db_expenses[db_expense_itr].Date.Year()

        // Last iteration
        if db_expense_itr + 1 == len(db_expenses) {
            api_expenses = append(api_expenses, api_expenses_year)
        }
    }

    api_result, err := json.Marshal(api_expenses)
    if err != nil {
        log.Fatal(err)
    }

    return string(api_result)
}

func(self *Index) db_set() string {
    // var local_data [1][3][12][5]int

    // for i := 0; i < 3; i++ {
    //     for j := 0; j < 12; j++ {
    //         for k := 0; k < 5; k++ {
    //             local_data[0][i][j][k] = k + 1
    //         }
    //     }
    // }

    // data := self.Data{Value: local_data}

    // err = collection.Insert(data)
    // if err != nil {
    //     log.Fatal(err)
    // }
    return "test"
}


func (self *Index) route(app *martini.ClassicMartini) {
    app.Get(
        "/",
        func(render render.Render) {
            render.JSON(
                200,
                map[string]interface{}{
                    "success": map[string]interface{}{"greeting": "Hello, I'm your API!"},
                    "error": nil,
                },
            )
        },
    )

    app.Post( // TODO: Second req — fail, not idempotent?
        "/api/v1/get",
        func(render render.Render) {
            render.JSON(200, self.db_get())
        },
    )

    app.Post(
        "/api/v1/set",
        func(render render.Render) {
            render.JSON(200, self.db_set())
        },
    )
}


func main() {
    martini_app := martini.Classic()
    martini_app.Use(render.Renderer())

    app := Index{}
    app.db_init()
    app.route(martini_app)

    fmt.Printf("App starting!\n")

    martini_app.Run()
}
