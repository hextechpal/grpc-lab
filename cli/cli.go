package cli

import (
	"github.com/ppal31/grpc-lab/cli/bookstore"
	lb "github.com/ppal31/grpc-lab/cli/lb/server"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var application = "Chat Application"
var description = "Chat Application description. Must some explaining text should be put here but I am super lazyy"

func Command() {
	app := kingpin.New(application, description)
	lb.Register(app)
	bookstore.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
