package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./openapi"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Nexus API</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/openapi.swagger.json',
            dom_id: '#swagger-ui',
        })
    </script>
</body>
</html>
`
		w.Write([]byte(html))
	})

	http.Handle("/openapi.swagger.json", http.StripPrefix("/", fs))

	log.Println("Swagger UI running at http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
