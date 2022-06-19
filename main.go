package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

const fPermission = 0644

var operationError = (errors.New("-operation flag has to be specified"))
var fileNameError = (errors.New("-fileName flag has to be specified"))
var itemError = (errors.New("-item flag has to be specified"))
var idError = (errors.New("-id flag has to be specified"))
var removeIdError = (errors.New("-id flag has to be specified"))

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func Perform(args Arguments, writer io.Writer) error {
	if err := validationFlagsOnEmpty(args["operation"], operationError); err != nil {
		return err
	}
	if err := validationFlagsOnEmpty(args["fileName"], fileNameError); err != nil {
		return err
	}
	if err := validationFlagsOnEmpty(args["id"], idError); args["operation"] == "findById" && err != nil {
		return err
	}
	if err := validationFlagsOnEmpty(args["id"], removeIdError); args["operation"] == "remove" && err != nil {
		return err
	}
	if err := validationFlagsOnEmpty(args["item"], itemError); args["operation"] == "add" && err != nil {
		return err
	}

	f, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, fPermission)

	if err != nil {
		return err
	}
	defer f.Close()
	ctx, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}
	var users []User

	switch args["operation"] {
	case "list":
		{
			err2 := json.Unmarshal(ctx, &users)
			if err2 != nil {
				return err2
			}
			writer.Write(ctx)
		}
	case "findById":
		{
			err2 := json.Unmarshal(ctx, &users)
			if err2 != nil {
				return err2
			}
			id := args["id"]
			if len(users) > 0 {
				for _, val := range users {
					if val.Id == id {
						u, err := json.Marshal(val)
						if err != nil {
							panic(err)
						}
						writer.Write(u)
					}
				}
			}
			break
		}
	case "add":
		{
			if len(ctx) > 0 {
				err2 := json.Unmarshal(ctx, &users)
				if err2 != nil {
					return err2
				}
			}
			var user User

			err1 := json.Unmarshal([]byte(args["item"]), &user)
			if err1 != nil {
				return err1
			}
			for _, val := range users {
				if val.Id == user.Id {
					s := fmt.Sprintf("Item with id %s already exists", user.Id)
					writer.Write([]byte(s))
					return nil
				}
			}
			users = append(users, user)

			SaveUser(users, args["fileName"])

			file, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, fPermission)

			if err != nil {
				return err
			}
			b, _ := ioutil.ReadAll(file)
			writer.Write(b)
			break
		}
	case "remove":
		{
			err2 := json.Unmarshal(ctx, &users)
			if err2 != nil {
				return err2
			}
			id := args["id"]

			var notFoundFlag bool = true

			for i, val := range users {
				if val.Id == id {
					users = append(users[:i], users[i+1:]...)
					notFoundFlag = false
				}
			}

			if notFoundFlag {
				s := fmt.Sprintf("Item with id %s not found", id)
				writer.Write([]byte(s))
				return nil

			}
			SaveUser(users, args["fileName"])
			break
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])

	}
	return nil
}

func parseArgs() Arguments {
	flag.String("operation", "", "Setting operation flag")
	flag.String("fileName", "", "Setting fileName flag")
	flag.String("item", "", "Setting item flag")
	flag.String("id", "", "Setting id flag")

	flag.Parse()

	argsMap := make(Arguments)
	args := os.Args[1:]
	for i := 0; i < len(args)-1; i += 2 {
		s := strings.Replace(args[i], "-", "", 1)
		argsMap[s] = args[i+1]
	}
	return argsMap
}

func validationFlagsOnEmpty(flagName string, err error) error {
	if flagName == "" {
		return err
	}
	return nil
}

func SaveUser(users []User, fileName string) {
	js, err := json.Marshal(&users)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(fileName, js, fPermission)
}
