package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

///////////////////////////////////////////////////////////////////////////////

func UUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	uuid[8] = 0x80
	uuid[4] = 0x40
	boo := hex.EncodeToString(uuid)
	return boo, nil
}

///////////////////////////////////////////////////////////////////////////////

func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func InsertOne(client *mongo.Client, ctx context.Context, dataBase, col string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := client.Database(dataBase).Collection(col)
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

func UpdateOne(client *mongo.Client, ctx context.Context, filter interface{}, dataBase string, col string, update interface{}) (*mongo.UpdateResult, error) {
	collection := client.Database(dataBase).Collection(col)
	result, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return result, err
}

func Query(client *mongo.Client, ctx context.Context, dataBase, col string, query, field interface{}) (result *mongo.Cursor, err error) {
	collection := client.Database(dataBase).Collection(col)
	result, err = collection.Find(ctx, query, options.Find().SetProjection(field))
	return
}

func CheckError(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		log.Println(msg)
		log.Println(err)
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func ShowIndex(w http.ResponseWriter, r *http.Request) {
	tmppath := "./static/index.html"
	tmpl := template.Must(template.ParseFiles(tmppath))
	tmpl.Execute(w, tmpl)
}

func ShowAdmin(w http.ResponseWriter, r *http.Request) {
	showtmppath := "./static/admin.html"
	showtmpl := template.Must(template.ParseFiles(showtmppath))
	showtmpl.Execute(w, showtmpl)
}

func AlphaT_Insert(db string, coll string, ablob ReviewStruct) {
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	CheckError(err, "AlphaT_Insert_: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, db, coll, ablob)
	CheckError(err2, "AlphaT_Insert_has failed")
}

func AlphaT_Insert_Pics(db string, coll string, picinfo PicStruct) {
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	CheckError(err, "AlphaT_Insert_Pics: Connections has failed")
	defer Close(client, ctx, cancel)
	_, err2 := InsertOne(client, ctx, db, coll, picinfo)
	CheckError(err2, "AlphaT_Insert_Picshas failed")
}

func AddToQuarantineHandler(w http.ResponseWriter, r *http.Request) {
	uuid, _ := UUID()
	var name string = r.URL.Query().Get("name")
	var email string = r.URL.Query().Get("email")
	var message string = r.URL.Query().Get("message")
	var sig string
	if name != "" {
		sig = name
	} else if email != "" {
		s := strings.Split(email, "@")
		sig = s[0]
	} else {
		sig = ""
	}

	ct := time.Now()
	date := ct.Format("01-01-2021")

	var newReview = ReviewStruct{
		UUID:       uuid,
		Date:       date,
		Name:       name,
		Email:      email,
		Sig:        sig,
		Message:    message,
		Approved:   "no",
		Quarintine: "yes",
		Delete:     "no",
	}
	AlphaT_Insert("maindb", "main", newReview)
	m1 := "<p>A new review was posted</p>"
	m2 := "<a href='http://34.127.50.188/Admin'>AlphaTreeService Admin Page</>"
	m3 := m1 + m2
	m := gomail.NewMessage()
	m.SetHeader("From", "porthose.cjsmo.cjsmo@gmail.com")
	m.SetHeader("To", "porthose.cjsmo.cjsmo@gmail.com", "Alpha.treeservicecdm@gmail.com")
	m.SetHeader("Subject: NEW REVIEW Has Been Posted")
	m.SetBody("text/html", m3)
	d := gomail.NewDialer("smtp.gmail.com", 587, "porthose.cjsmo.cjsmo@gmail.com", "!Porthose1960")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func AllQuarintineReviewsHandler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{"approved": "no", "quarintine": "yes", "delete": "no"}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("maindb").Collection("main")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllQuarintineReviews find has failed")
	var allQRevs []ReviewStruct
	if err = cur.All(context.TODO(), &allQRevs); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s this is AllQuarintineReviews-", allQRevs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&allQRevs)
}

func AllApprovedReviewsHandler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{"approved": "yes", "quarintine": "no", "delete": "no"}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("maindb").Collection("main")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllReviews find has failed")
	var allRevs []ReviewStruct
	if err = cur.All(context.TODO(), &allRevs); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s this is AllReviews-", allRevs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&allRevs)
	log.Println("AllReviews Info Complete")
}

func SetReviewToDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var delUUID string = r.URL.Query().Get("uuid")
	filter := bson.M{"uuid": delUUID}
	update := bson.M{"$set": bson.M{"delete": "yes"}}
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	UpdateOne(client, ctx, filter, "maindb", "main", update)
}

func ProcessQuarantineHandler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("maindb").Collection("main")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllQuarintineReviews find has failed")
	var allRevs []ReviewStruct
	if err = cur.All(context.TODO(), &allRevs); err != nil {
		log.Fatal(err)
	}
	for _, rev := range allRevs {
		filter := bson.M{"uuid": rev.UUID}
		update := bson.M{"$set": bson.M{"approved": "yes", "quarintine": "no"}}
		client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
		defer Close(client, ctx, cancel)
		CheckError(err, "MongoDB connection has failed")
		UpdateOne(client, ctx, filter, "maindb", "main", update)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Update complete")
	log.Println("AllQuarintineReviews Info Complete")
}

func BackupReviewHandler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("maindb").Collection("main")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllReviews find has failed")
	var allRevs []ReviewStruct
	if err = cur.All(context.TODO(), &allRevs); err != nil {
		log.Fatal(err)
	}
	bString, _ := json.Marshal(allRevs)
	err = ioutil.WriteFile("/root/backup/backup.json", bString, 0644)
	if err != nil {
		log.Fatal(err)
	}
	name_of_file := "backup.json"
	f, _ := os.Open("/root/backup/" + name_of_file)
	read := bufio.NewReader(f)
	data, _ := ioutil.ReadAll(read)
	name_of_file = strings.Replace(name_of_file, ".json", ".gz", -1)
	f, _ = os.Create("/root/backup/" + name_of_file)
	ww := gzip.NewWriter(f)
	ww.Write(data)
	ww.Close()
	t := time.Now().Format(time.RFC3339)
	tstring := string(t)
	s := "<p>AlphaTreeService Reviews Backup for: " + tstring + "</p>"
	fmt.Println("this is s")
	fmt.Println(s)
	m := gomail.NewMessage()
	m.SetHeader("From", "porthose.cjsmo.cjsmo@gmail.com")
	m.SetHeader("To", "porthose.cjsmo.cjsmo@gmail.com", "Alpha.treeservicecdm@gmail.com")
	m.SetHeader("Subject: AlphaTreeService Reviews Backup")
	m.SetBody("text/html", s)
	m.Attach("/root/backup/" + name_of_file)
	d := gomail.NewDialer("smtp.gmail.com", 587, "porthose.cjsmo.cjsmo@gmail.com", "!Porthose1960")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func ShowGalleryPage1Handler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{"page": "1"}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("picdb").Collection("portrait")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllQuarintineReviews find has failed")
	var allPage1 []PicStruct
	if err = cur.All(context.TODO(), &allPage1); err != nil {
		log.Fatal(err)
	}
	tmpl2 := template.Must(template.ParseFiles("./static/gallery.html"))
	tmpl2.Execute(w, allPage1)
}

func ShowGalleryPage2Handler(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{"page": "2"}
	opts := options.Find()
	opts.SetProjection(bson.M{"_id": 0})
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "MongoDB connection has failed")
	coll := client.Database("picdb").Collection("landscape")
	cur, err := coll.Find(context.TODO(), filter, opts)
	CheckError(err, "AllQuarintineReviews find has failed")
	var allPage2 []PicStruct
	if err = cur.All(context.TODO(), &allPage2); err != nil {
		log.Fatal(err)
	}
	tmpl2 := template.Must(template.ParseFiles("./static/gallery.html"))
	tmpl2.Execute(w, allPage2)
}

func AtsGoFindOnePic(db string, coll string, filtertype string, filterstring string) PicStruct {
	filter := bson.M{filtertype: filterstring}
	client, ctx, cancel, err := Connect("mongodb://db:27017/atsgodb")
	defer Close(client, ctx, cancel)
	CheckError(err, "AtsGoFindOnePic: MongoDB connection has failed")
	collection := client.Database(db).Collection(coll)
	var results PicStruct
	err = collection.FindOne(context.Background(), filter).Decode(&results)
	if err != nil {
		log.Println("AtsGoFindOnePic: find one has fucked up")
		log.Fatal(err)
	}
	return results
}

func ZoomPic1Handler(w http.ResponseWriter, r *http.Request) {
	// portrait
	pid := r.URL.Query().Get("picid")
	fmt.Println(pid)
	pic := AtsGoFindOnePic("picdb", "portrait", "picid", pid)
	fmt.Println(pic)
	tmpl2 := template.Must(template.ParseFiles("./static/zoom.html"))
	tmpl2.Execute(w, pic)
}

func ZoomPic2Handler(w http.ResponseWriter, r *http.Request) {
	// landscape
	picid2 := r.URL.Query().Get("picid")
	fmt.Println(picid2)
	pic2 := AtsGoFindOnePic("picdb", "landscape", "picid", picid2)
	fmt.Println(pic2)
	tmpl2 := template.Must(template.ParseFiles("./static/zoom.html"))
	tmpl2.Execute(w, pic2)
}

type ReviewStruct struct {
	UUID       string `yaml:"UUID"`
	Date       string `yaml:"Date"`
	Name       string `yaml:"Name"`
	Email      string `yaml:"Email"`
	Sig        string `yaml:"Sig"`
	Message    string `yaml:"Message"`
	Approved   string `yaml:"Approved"`
	Quarintine string `yaml:"Quarintine"`
	Delete     string `yaml:"Delete"`
}

func (c *ReviewStruct) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

type PicStruct struct {
	PicID  string `bson:"picid"`
	Pic    string `bson:"pic"`
	Thumb  string `bson:"thumb"`
	Page   string `bson:"page"`
	Orient bool   `bson:"orient"`
}

func init() {
	data, err := ioutil.ReadFile("./static/review1.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var rev1 ReviewStruct
	if err := rev1.Parse(data); err != nil {
		log.Fatal(err)
	}
	fmt.Println(rev1)
	AlphaT_Insert("maindb", "main", rev1)

	data2, err := ioutil.ReadFile("./static/review2.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var rev2 ReviewStruct
	if err := rev2.Parse(data2); err != nil {
		log.Fatal(err)
	}
	fmt.Println(rev2)
	AlphaT_Insert("maindb", "main", rev2)

	data3, err := ioutil.ReadFile("./static/fake1.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var rev3 ReviewStruct
	if err := rev3.Parse(data3); err != nil {
		log.Fatal(err)
	}
	fmt.Println(rev3)
	AlphaT_Insert("maindb", "main", rev3)

	data4, err := ioutil.ReadFile("./static/fake2.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var rev4 ReviewStruct
	if err := rev4.Parse(data4); err != nil {
		log.Fatal(err)
	}
	fmt.Println(rev4)
	AlphaT_Insert("maindb", "main", rev4)

	g1, _ := filepath.Glob("./static/gallery/landscape/*.webp")

	countPage := 1
	for idx, g := range g1 {
		if strings.Contains(g, "_thumb") {
			println("this is thumb")
		} else {
			var newpic PicStruct
			if strings.Contains(g, "landscape") {
				newpic.Orient = false // "landscape"
			} else {
				newpic.Orient = true //"portrait"
			}
			newpic.PicID, _ = UUID()
			newpic.Pic = os.Getenv("ATSGO_SERVER_ADDR") + "/" + g
			fmt.Println(newpic.Pic)
			ext := filepath.Ext(g)
			newpic.Thumb = os.Getenv("ATSGO_SERVER_ADDR") + "/" + g[:len(g)-5] + "_thumb" + ext
			fmt.Println(newpic.Thumb)
			if idx%50 == 0 {
				countPage += 1
				newpic.Page = strconv.Itoa(countPage)
				AlphaT_Insert_Pics("picdb", "landscape", newpic)
			} else {
				newpic.Page = strconv.Itoa(countPage)
				AlphaT_Insert_Pics("picdb", "landscape", newpic)
			}
		}
	}

	g2, _ := filepath.Glob("./static/gallery/portrait/*.webp")
	countPage2 := 0
	for idx, gg := range g2 {
		if strings.Contains(gg, "_thumb") {
			println("this is thumb")
		} else {
			var newpic PicStruct
			if strings.Contains(gg, "landscape") {
				newpic.Orient = false //"landscape"
			} else {
				newpic.Orient = true //"portrait"
			}
			newpic.PicID, _ = UUID()
			newpic.Pic = "./" + gg
			fmt.Println(newpic.Pic)
			ext := filepath.Ext(gg)
			newpic.Thumb = "./" + gg[:len(gg)-5] + "_thumb" + ext
			fmt.Println(newpic.Thumb)
			if idx%50 == 0 {
				countPage2 += 1
				newpic.Page = strconv.Itoa(countPage2)
				AlphaT_Insert_Pics("picdb", "portrait", newpic)
			} else {
				newpic.Page = strconv.Itoa(countPage2)
				AlphaT_Insert_Pics("picdb", "portrait", newpic)
			}
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", ShowIndex)
	r.HandleFunc("/admin", ShowAdmin)
	r.HandleFunc("/galleryp1", ShowGalleryPage1Handler)
	r.HandleFunc("/galleryp2", ShowGalleryPage2Handler)
	r.HandleFunc("/zoompic1", ZoomPic1Handler)
	r.HandleFunc("/zoompic2", ZoomPic2Handler)
	r.HandleFunc("/AllQReviews", AllQuarintineReviewsHandler)
	r.HandleFunc("/AllApprovedReviews", AllApprovedReviewsHandler)
	r.HandleFunc("/ProcessQuarintine", ProcessQuarantineHandler)
	r.HandleFunc("/Backup", BackupReviewHandler)
	r.HandleFunc("/DeleteReview", SetReviewToDeleteHandler)
	r.HandleFunc("/atq", AddToQuarantineHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	port := ":80"
	http.ListenAndServe(port, (r))
}
