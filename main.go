package main

import(
	"github.com/gin-gonic/gin"
	"time"
	"gopkg.in/mgo.v2"
	_ "gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/bson"
)
type MongoDB struct {
	Host             string
	Port             string
	Addrs            string
	Database         string
	EventTTLAfterEnd time.Duration
	StdEventTTL      time.Duration
	Info             *mgo.DialInfo
	Session          *mgo.Session
}
type Data struct {
	Id   bson.ObjectId `form:"id" bson:"_id,omitempty"`
	Data string        `form:"data" bson:"data"`
}

func (mongo *MongoDB) SetDefault() { // {{{
	mongo.Host = "localhost"
	mongo.Addrs = "localhost:8000"
	mongo.Database = "context"
	mongo.Info = &mgo.DialInfo{
		Addrs:    []string{mongo.Addrs},
		Database: mongo.Database,
	}
}

func main() {

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	mongo := MongoDB{}
	mongo.SetDefault()
	router.Use(MiddleDB(&mongo))

	router.GET("/", func(context *gin.Context) {
		context.String(200,"hello world")
		})
	//r.Use(middlewares.DB())

	router.GET("/", contactList)
	router.POST("/", addressBookView)

	router.Run(":8000")
}


func MiddleDB(mongo *MongoDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := mongo.SetSession()
		if err != nil {
			c.Abort()
		} else {
			c.Set("mongo", mongo)
			c.Next()
		}
	}
}
func (mongo *MongoDB) SetSession() (err error) {
	mongo.Session, err = mgo.DialWithInfo(mongo.Info)
	if err != nil {
		mongo.Session, err = mgo.Dial(mongo.Host)
		if err != nil {
			return err
		}
	}
	return err
}

func contactList(c *gin.Context) {
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't reach db", "body": nil})
	}

	data, err := mongo.GetData()
	// fmt.Printf("\ndata: %v, ok: %v\n", data, ok)
	if err != nil {
		c.JSON(400, gin.H{"message": "can't get data from database", "body": nil})
	} else {
		c.JSON(200, gin.H{"message": "get data sucess", "body": data})
	}
}

func (mongo *MongoDB) GetData() (dates []Data, err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Data").Find(bson.M{}).All(&dates)
	return dates, err
}

func addressBookView(c *gin.Context)  {
	mongo, ok := c.Keys["mongo"].(*MongoDB)
	if !ok {
		c.JSON(400, gin.H{"message": "can't connect to db", "body": nil})
	}
	var req Data
	err := c.Bind(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": "Incorrect data", "body": nil})
		return
	} else {
		err := mongo.PostData(&req)
		if err != nil {
			c.JSON(400, gin.H{"message": "error post to db", "body": nil})
		}
		c.JSON(200, gin.H{"message": "post data sucess", "body": req})
	}
}
func (mongo *MongoDB) PostData(data *Data) (err error) { // {{{
	session := mongo.Session.Clone()
	defer session.Close()

	err = session.DB(mongo.Database).C("Data").Insert(&data)
	return err
}
