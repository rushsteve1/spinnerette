(import spin/responses :as r)

(r/json
 {:title "Supported response types"
  :items ["JSON via spin/json"
          "HTML via janet-html, which is built-in"
          "strings, as the return value of a janet script"
          "Circlet response objects"
          :keywords-become-strings]}
 # Supports some optional parameters like http-response status
 :status 201)
