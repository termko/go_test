package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

type histMsg struct {
	Msg map[string]string `json:"msg"`
}

type lineMsg struct {
	Msg string `json:"msg"`
}

type echoMsg struct {
	Msg string `json:"msg"`
}

type key int

const login key = 0

func historyFromRedis(user string) (map[string]string, error) {
	redisConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return nil, err
	}
	defer redisConn.Close()

	ans, err := redis.StringMap(redisConn.Do("HGETALL", "user:"+user))
	if err != nil {
		return nil, err
	}
	fmt.Println("Done fetching history from redis!")
	return ans, nil
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	lines, ok := vals["line"]
	if ok {
		if len(lines) >= 1 {
			lineHandler(w, r, lines[0])
			return
		}
	}
	fmt.Println("Got to historyHandler!")

	ctx := r.Context()
	logpass, ok := ctx.Value(login).(string)
	if !ok {
		fmt.Println("Error getting username:", logpass)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong..."))
		return
	}
	givenLogPass := strings.Split(logpass, ":")
	givenLogin := givenLogPass[0]

	v, err := historyFromRedis(givenLogin)
	if err != nil {
		fmt.Println("Couldn't return history from redis: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong..."))
		return
	}

	resp := histMsg{
		Msg: v,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Write(data)
}

func lineFromRedis(line string, user string) (string, error) {
	redisConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return "", err
	}
	defer redisConn.Close()

	ans, err := redis.String(redisConn.Do("HGET", "user:"+user, line))
	if err != nil {
		return "", err
	}
	fmt.Println("Done fetching history from redis!")
	return ans, nil
}

func lineHandler(w http.ResponseWriter, r *http.Request, lineNumber string) {
	ctx := r.Context()
	logpass, ok := ctx.Value(login).(string)
	if !ok {
		fmt.Println("Error getting username:", logpass)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong..."))
		return
	}
	givenLogPass := strings.Split(logpass, ":")
	givenLogin := givenLogPass[0]

	line, err := lineFromRedis(lineNumber, givenLogin)
	if err != nil {
		if err.Error() != "redigo: nil returned" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something went wrong"))
			fmt.Println(err.Error(), "Redis")
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No such line index"))
		fmt.Println("Wrong line index for", givenLogin, ":", lineNumber, err.Error())
		return
	}
	resp := lineMsg{
		Msg: line,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		fmt.Println(err.Error(), "Json")
		return
	}

	w.Write(data)
}

func echoToRedis(user string, msg string) error {
	redisConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}
	defer redisConn.Close()

	ans, err := redisConn.Do("HLEN", "user:"+user)
	if err != nil {
		return err
	}

	_, err = redisConn.Do("HMSET", "user:"+user, ans, msg)
	if err != nil {
		return err
	}
	fmt.Println("Done echoing to redis!")
	return nil
}

func echoHanlder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logpass, ok := ctx.Value(login).(string)
	if !ok {
		fmt.Println("Error getting username:", logpass)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong..."))
		return
	}
	givenLogPass := strings.Split(logpass, ":")
	givenLogin := givenLogPass[0]
	// tmp := strings.Split(string(r.Header.Get("Authorization")), " ")
	// logpass, err := base64.StdEncoding.DecodeString(tmp[1])
	// if err != nil {
	// 	fmt.Println("Couldn't decode username: ", err.Error())
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte("Something went wrong..."))
	// 	return
	// }
	// log := strings.Split(string(logpass), ":")

	var v echoMsg
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		fmt.Println("Couldn't decode request body json: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong request body json"))
		return
	}
	if v.Msg == "" {
		fmt.Println("Missing request body json Msg:", v.Msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing request body json Msg"))
		return
	}

	err = echoToRedis(givenLogin, v.Msg)
	if err != nil {
		fmt.Println("Couldn't echo to redis: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong..."))
		return
	}

	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Write(data)
}

func mw(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if !strings.HasPrefix(header, "Basic ") {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Wrong username or password\n"))
				fmt.Println("Someone tried to login with wrong credentials")
				return
			}
			loginPassword, err := base64.StdEncoding.DecodeString(header[6:])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Wrong username or password\n"))
				fmt.Println("Someone tried to login with wrong credentials")
				return
			}
			fmt.Println(string(loginPassword))
			ctx := context.Background()
			ctx = context.WithValue(ctx, login, strings.Split(string(loginPassword), ":")[0])
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		},
	)
}

func readLogPassFile() {
	redisConn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer redisConn.Close()

	if len(os.Args) != 2 {
		log.Fatal("Usage: program <files>\n")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		_, err = redisConn.Do("SADD", "logpass", scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading logpass file:", err)
		os.Exit(1)
	}

	fmt.Println("LogPass Parsed!")
}

func main() {
	readLogPassFile()

	server := http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	r := mux.NewRouter()

	r.Use(mw)
	r.HandleFunc("/echo", echoHanlder).Methods(http.MethodPost)
	r.HandleFunc("/history", historyHandler).Methods(http.MethodGet)

	server.Handler = r

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
