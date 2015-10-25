package main

import ("fmt"
		"net/http"
		"encoding/json"
		"io/ioutil"
		"gopkg.in/mgo.v2"
        "gopkg.in/mgo.v2/bson"
        "log"
        "github.com/julienschmidt/httprouter"
        "strconv"
        "strings"
        "time"
)

const(
	MongoDBHosts = "ds045454.mongolab.com:45454" 
 	AuthDatabase = "rugved" 
 	AuthUserName = "rugved" 
 	AuthPassword = "rugved" 
)

type Response struct{
	Id int `json:"id" bson:"id"` 
	Name string `json:"name" bson:"name"`
	Address string `json:"address" bson:"address"`
	City string `json:"city" bson:"city"`
	State string `json:"state" bson:"state"`
	Zip string `json:"zip" bson:"zip"`
	Coordin Coordinate `json:"coordinate" bson:"coordinate"`
}

type Coordinate struct{
	Latitude float64 `json:"lat" bson:"lat"`
	Longitude float64 `json:"lng" bson:"lng"`
}

type Result struct{
	Id int  `bson:"id"`
}

func postLocation(respwriter http.ResponseWriter, request *http.Request,p httprouter.Params) {
	
	var idResult Result
	dbConnection := &mgo.DialInfo{ 
	Addrs:    []string{MongoDBHosts}, 
 	Timeout:  60 * time.Second, 
 	Database: AuthDatabase, 
 	Username: AuthUserName, 
 	Password: AuthPassword, 
	} 
	sesn, err := mgo.DialWithInfo(dbConnection)
	if(err!=nil){
		defer sesn.Close()
	}
	var jsonLocatn interface{}
	var data Response
	rq,err:= ioutil.ReadAll(request.Body)
	json.Unmarshal(rq,&data)
	addr:= strings.Replace(data.Address," ","+",-1)
	addr= addr + ",+"+ strings.Replace(data.City," ","+",-1)
	addr= addr + ",+" + data.State
	response,err:= http.Get("http://maps.google.com/maps/api/geocode/json?address="+addr+"&sensor=false")
	if err!=nil{
		fmt.Println("Error:",err)
	}else{
		defer response.Body.Close()
		locationContent,_:= ioutil.ReadAll(response.Body)
		json.Unmarshal(locationContent,&jsonLocatn)
		latitude:= (jsonLocatn.(map[string]interface{})["results"]).([]interface{})[0].(map[string]interface{})["geometry"].
		           (map[string]interface{})["location"].(map[string]interface{})["lat"]
		longitude:= (jsonLocatn.(map[string]interface{})["results"]).([]interface{})[0].(map[string]interface{})["geometry"].
		           (map[string]interface{})["location"].(map[string]interface{})["lng"]
		data.Coordin.Latitude=latitude.(float64)
		data.Coordin.Longitude=longitude.(float64)
		con := sesn.DB("rugved").C("assignment2")
		idResult.Id=0
		count,_:=con.Count()
		if(count > 0){
			err := con.Find(nil).Select(bson.M{"id":1}).Sort("-id").One(&idResult)
			if(err!=nil){
				log.Printf("RunQuery : ERROR : %s\n", err) 
				fmt.Fprintln(respwriter,err)
				return 
			}
			data.Id = idResult.Id + 1
	        err = con.Insert(data)
	        if err != nil {
	                log.Fatal(err)
	        }
	        result,_:=json.Marshal(data)
			fmt.Fprintln(respwriter,string(result))
		}else{
			data.Id = idResult.Id + 1
	        err = con.Insert(data)
	        if err != nil {
	                log.Fatal(err)
	        }
	        result,_:=json.Marshal(data)
			fmt.Fprintln(respwriter,string(result))	
		}
	}
}

func getLocation(respwriter http.ResponseWriter, request *http.Request,p httprouter.Params){
	params,_:= strconv.Atoi(p.ByName("locationid"))
	dbConnection := &mgo.DialInfo{ 
	Addrs:    []string{MongoDBHosts}, 
 	Timeout:  60 * time.Second, 
 	Database: AuthDatabase, 
 	Username: AuthUserName, 
 	Password: AuthPassword, 
	} 
	sesn, err := mgo.DialWithInfo(dbConnection)
	if(err!=nil){
		defer sesn.Close()
	}
	var data Response
	con := sesn.DB("rugved").C("assignment2")
	err = con.Find(bson.M{"id":params}).Select(bson.M{"_id":0}).One(&data)
	if(err!=nil){
		log.Printf("RunQuery : ERROR : %s\n", err) 
		fmt.Fprintln(respwriter,err)
				return
	}else{
		result,_:=json.Marshal(data)
		fmt.Fprintln(respwriter,string(result))	
	}
}

func putLocation(respwriter http.ResponseWriter, request *http.Request,p httprouter.Params){
	rq,err:= ioutil.ReadAll(request.Body)
	params,_:= strconv.Atoi(p.ByName("locationid"))
	dbConnection := &mgo.DialInfo{ 
	Addrs:    []string{MongoDBHosts}, 
 	Timeout:  60 * time.Second, 
 	Database: AuthDatabase, 
 	Username: AuthUserName, 
 	Password: AuthPassword, 
	} 
	sesn, err := mgo.DialWithInfo(dbConnection)
	if(err!=nil){
		defer sesn.Close()
	}
	var data Response
	con := sesn.DB("rugved").C("assignment2")	
	json.Unmarshal(rq,&data)
	addr:= strings.Replace(data.Address," ","+",-1)
	addr= addr + ",+"+ strings.Replace(data.City," ","+",-1)
	addr= addr + ",+" + data.State
	var jsonLocatn interface{}
	response,err:= http.Get("http://maps.google.com/maps/api/geocode/json?address="+addr+"&sensor=false")
	if err!=nil{
		fmt.Println("Error:",err)
	}else{
		defer response.Body.Close()
		locationContent,_:= ioutil.ReadAll(response.Body)
		json.Unmarshal(locationContent,&jsonLocatn)		
		latitude:= (jsonLocatn.(map[string]interface{})["results"]).([]interface{})[0].(map[string]interface{})["geometry"].
		           (map[string]interface{})["location"].(map[string]interface{})["lat"]
		longitude:= (jsonLocatn.(map[string]interface{})["results"]).([]interface{})[0].(map[string]interface{})["geometry"].
		           (map[string]interface{})["location"].(map[string]interface{})["lng"]		
		data.Coordin.Latitude=latitude.(float64)
		data.Coordin.Longitude=longitude.(float64)
		err = con.Update(bson.M{"id":params},bson.M{"$set":bson.M{"address":data.Address,"city":data.City,"state":data.State,"zip":data.Zip,"coordinate.lat":data.Coordin.Latitude,"coordinate.lng":data.Coordin.Longitude}})
		if(err!=nil){
			log.Printf("RunQuery : ERROR : %s\n", err) 
			fmt.Fprintln(respwriter,err)
					return
		}else{
			err = con.Find(bson.M{"id":params}).Select(bson.M{"_id":0}).One(&data)
			if(err!=nil){
				log.Printf("RunQuery : ERROR : %s\n", err) 
				fmt.Fprintln(respwriter,err)
					return
			}
			result,_:=json.Marshal(data)
			fmt.Fprintln(respwriter,string(result))	
		}
	}
}

func deleteLocation(respwriter http.ResponseWriter, request *http.Request,p httprouter.Params){
	params,_:= strconv.Atoi(p.ByName("locationid"))
	dbConnection := &mgo.DialInfo{ 
	Addrs:    []string{MongoDBHosts}, 
 	Timeout:  60 * time.Second, 
 	Database: AuthDatabase, 
 	Username: AuthUserName, 
 	Password: AuthPassword, 
	} 
	sesn, err := mgo.DialWithInfo(dbConnection)
	if(err!=nil){
		defer sesn.Close()
	}
	con := sesn.DB("rugved").C("assignment2")	
	err = con.Remove(bson.M{"id":params})
	if(err!=nil){
		log.Printf("RunQuery : ERROR : %s\n", err) 
		fmt.Fprintln(respwriter,err)
				return
	}else{
		fmt.Fprintln(respwriter,"Deleted a record")
	}
}

func main() {
	mux := httprouter.New()
    mux.POST("/locations",postLocation)
    mux.GET("/locations/:locationid",getLocation)
    mux.PUT("/locations/:locationid",putLocation)
    mux.DELETE("/locations/:locationid",deleteLocation)
     srvr := http.Server{
            Addr:        "0.0.0.0:8086",
            Handler: mux,
    }
    srvr.ListenAndServe()
}