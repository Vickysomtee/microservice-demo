package cmd

import (
	"log"
	"os"

	gw "github.com/Joker666/microservice-demo/protos/api"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"context"
	"flag"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// proxyCmd is the serve sub command to start the api server
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "proxy tests the api server",
	RunE:  proxy,
}

func init() {
	//
}

func proxy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := mux.NewRouter()
	g, err := newGateway(ctx)
	if err != nil {
		return err
	}
	mux.Handle("/", g)

	mux.HandleFunc("/swagger.json", serveSwagger)
	fs := http.FileServer(http.Dir("www/swagger-ui"))
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui", fs))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	proxyPort := ":" + os.Getenv("PROXY_PORT")
	log.Println("Starting proxy server at " + proxyPort)
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(proxyPort, c.Handler(mux))
}

func newGateway(ctx context.Context) (http.Handler, error) {
	grpcServerAddress := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	grpcServerEndpoint := flag.String("grpc-server-endpoint", grpcServerAddress, "gRPC server endpoint")
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "www/api.swagger.json")
}
