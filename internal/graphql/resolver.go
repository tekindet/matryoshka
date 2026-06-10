package graphql

import (
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/tekindet/matryoshka/internal/manager"
)

type Resolver struct {
	mgr manager.Manager
}

func New(mgr manager.Manager) *Resolver {
	return &Resolver{mgr: mgr}
}

func (r *Resolver) Handler() http.Handler {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    r.QueryType(),
		Mutation: r.MutationType(),
	})
	if err != nil {
		panic(err)
	}

	return handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
}

var projectType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Project",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"description": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var serviceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Service",
	Fields: graphql.Fields{
		"id":           &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"projectId":    &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
		"name":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"type":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"status":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"externalPort": &graphql.Field{Type: graphql.Int},
	},
})

func (r *Resolver) QueryType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"project": &graphql.Field{
				Type: projectType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.mgr.GetProject(p.Context, p.Args["id"].(string))
				},
			},
			"projects": &graphql.Field{
				Type: graphql.NewList(graphql.NewNonNull(projectType)),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.mgr.ListProjects(p.Context)
				},
			},
		},
	})
}

func (r *Resolver) MutationType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createProject": &graphql.Field{
				Type: projectType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
							Name: "CreateProjectInput",
							Fields: graphql.InputObjectConfigFieldMap{
								"name":        &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
								"description": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
							},
						})),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					return r.mgr.CreateProject(p.Context, input["name"].(string), input["description"].(string))
				},
			},
			"createService": &graphql.Field{
				Type: serviceType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
							Name: "CreateServiceInput",
							Fields: graphql.InputObjectConfigFieldMap{
								"projectId": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.ID)},
								"name":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
								"type":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
							},
						})),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					return r.mgr.CreateService(p.Context, input["projectId"].(string), input["name"].(string), input["type"].(string))
				},
			},
		},
	})
}
