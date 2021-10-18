(import json)

(defn json
  "
  Create a JSON response
  Takes serializable content
  Returns a Circle response map with the appropriate headers
  "
  [content &opt &keys {:status status}]
  {:headers {"Content-Type" "application/json"}
   :status (or status 200)
   :body (json/encode content)})
