package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
	uuid "github.com/satori/go.uuid"
	gocb "gopkg.in/couchbase/gocb.v1"
	"github.com/friendsofgo/graphiql"
)

type Account struct {
	ID        string `json:"id,omitempty"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Type      string `json:"type"`
}

type Blog struct {
	ID      string `json:"id,omitempty"`
	Account string `json:"account"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

var bucket *gocb.Bucket

func main() {
	fmt.Println("Starting Application")
	cluster, err := gocb.Connect("couchbase://localhost")
	if err != nil {
		log.Fatal(err)
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{Username: "root", Password: "123456"})
	bucket, err = cluster.OpenBucket("example", "")
	if err != nil {
		log.Fatal(err)
	}
	accountType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Account",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"firstname": &graphql.Field{
				Type: graphql.String,
			},
			"lastname": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	blogType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Blog",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"account": &graphql.Field{
				Type: graphql.String,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"content": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"accounts": &graphql.Field{
				Type: graphql.NewList(accountType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query := gocb.NewN1qlQuery("SELECT META(account).id, account.* FROM example AS account WHERE account.type = 'account'")
					rows, err := bucket.ExecuteN1qlQuery(query, nil)
					if err != nil {
						return nil, err
					}
					var accounts []Account
					var row Account
					for rows.Next(&row) {
						accounts = append(accounts, row)
					}
					return accounts, nil
				},
			},
			"account": &graphql.Field{
				Type: accountType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var account Account
					account.ID = params.Args["id"].(string)
					_, err := bucket.Get(account.ID, &account)
					if err != nil {
						return nil, err
					}
					return account, nil
				},
			},
			"blogs": &graphql.Field{
				Type: graphql.NewList(blogType),
				Args: graphql.FieldConfigArgument{
					"account": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					account := params.Args["account"].(string)
					query := gocb.NewN1qlQuery("SELECT META(blog).id, blog.* FROM example as blog WHERE blog.type = 'blog' AND blog.account = $1")
					var n1q1Params []interface{}
					n1q1Params = append(n1q1Params, account)
					rows, err := bucket.ExecuteN1qlQuery(query, n1q1Params)
					if err != nil {
						return nil, err
					}
					var blogs []Blog
					var row Blog
					for rows.Next(&row) {
						blogs = append(blogs, row)
					}
					return blogs, nil
				},
			},
		},
	})
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"createAccount": &graphql.Field{
				Type: accountType,
				Args: graphql.FieldConfigArgument{
					"firstname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"lastname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var account Account
					account.Firstname = params.Args["firstname"].(string)
					account.Lastname = params.Args["lastname"].(string)
					account.Type = "account"
					id := uuid.NewV4()
					_, err := bucket.Insert(id.String(), &account, 0)
					if err != nil {
						return nil, err
					}
					account.ID = id.String()
					return account, nil
				},
			},
			"updateAccount": &graphql.Field{
				Type: accountType,
				Args: graphql.FieldConfigArgument{
					"firstname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"lastname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"type": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var account Account
					account.Firstname = params.Args["firstname"].(string)
					account.Lastname = params.Args["lastname"].(string)
					account.Type = params.Args["type"].(string)
					id := params.Args["id"].(string)
					account.ID = id
					// Retrieve Document
					var retValue Account
					_, err = bucket.Get(id, &retValue)
					if err != nil {
						fmt.Println("ERROR RETURNING DOCUMENT:", err)
					}
					fmt.Println("----Document Retrieved----:", retValue)
					/// update ///
					// Replace the existing document
					_, err := bucket.Replace(id, &account, 0, 0)
					if err != nil {
						fmt.Println("ERROR REPLACING DOCUMENT:", err)
					}

					// Retrieve updated document
					_, err = bucket.Get(id, &account)
					if err != nil {
						fmt.Println("ERROR RETURNING DOCUMENT:", err)
					}
					fmt.Println("Document Retrieved:", account)

					// Exiting
					fmt.Println("Update Successful  - Exiting")
					return account, nil

					//// update end ///
				},
			},
			"deleteAccount": &graphql.Field{
				Type: accountType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var account Account
					account.ID = params.Args["id"].(string)
					_, err = bucket.Get(account.ID, &account)
					if err != nil {
						fmt.Println("ERROR RETURNING DOCUMENT:", err)
					}
					_, err := bucket.Remove(account.ID, 0)
					if err != nil {
						fmt.Println("ERROR REMOVING DOCUMENT:", err)
					}
					fmt.Println("DOCUMENT Removed:", account)
					return account, nil
				},
			},
			"createBlog": &graphql.Field{
				Type: blogType,
				Args: graphql.FieldConfigArgument{
					"account": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"title": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"content": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"type": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var blog Blog
					blog.Account = params.Args["account"].(string)
					blog.Title = params.Args["title"].(string)
					blog.Content = params.Args["content"].(string)
					blog.Type = params.Args["type"].(string)
					id := uuid.NewV4()
					_, err = bucket.Insert(id.String(), &blog, 0)
					if err != nil {
						return nil, err
					}
					blog.ID = id.String()
					return blog, nil
				},
			},
			"deleteBlog": &graphql.Field{
				Type: blogType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var blog Blog
					blog.ID = params.Args["id"].(string)
					_, err = bucket.Get(blog.ID, &blog)
					if err != nil {
						fmt.Println("ERROR RETURNING DOCUMENT:", err)
					}
					_, err := bucket.Remove(blog.ID, 0)
					if err != nil {
						fmt.Println("ERROR REMOVING DOCUMENT:", err)
					}
					return blog, nil
				},
			},
		},
	})

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"),
		})
		json.NewEncoder(w).Encode(result)
	})
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("http://localhost:3000/graphql")
	if err != nil {
    		panic(err)
	}
    	
	
	http.Handle("/graphiql", graphiqlHandler)
	http.ListenAndServe(":3000", nil)
}
