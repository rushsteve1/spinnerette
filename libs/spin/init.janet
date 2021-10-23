(import ./cache :export true)

(import json :as js)

(defn json
  "
  Create a JSON response
  Takes serializable content like a map, array, string, or number
  Returns a Circlet response map with the appropriate headers with the content
  encoded as a JSON string.
  "
  [content &opt &keys {:status status}]
  {:headers {"Content-Type" "application/json"}
   :status (or status 200)
   :body (js/encode content)})
