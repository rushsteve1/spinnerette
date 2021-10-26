(import json :as js)

(defn content-type
  "Creates an HTTP response with `Content-Type` set"
  [type content &opt &keys {:status status}]
  {:headers {"Content-Type" (string type)}
   :status (or status 200)
   :body (string content)})

(defn json
  `
  Create a JSON response
  Takes serializable content like a map, array, string, or number
  Returns a Circlet response map with the appropriate headers with the content
  encoded as a JSON string.
  `
  [content &opt &keys {:status status}]
  (content-type "application/json" (js/encode content) :status status))

(defn janet
  "Creates a marshalled Janet response"
  [content &opt &keys {:status status}]
  (content-type "text/janet" (marshal content) :status status))
