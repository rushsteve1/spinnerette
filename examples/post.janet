(import janet-html)

(def server-error-response
  {:status 500
   :body "Something went wrong, please try again later."})

(defn create-response
  [request]
  (janet-html/encode
    [:html
     [:body
      [:h1 "Post request example."]
      (if request
        [:p (string/format "Your request contained the following body: %s" (-> request (get :body) string))]
        [:p "Send a POST request here to see its content in the response."])]]))

(try
  {:status 200
   :body (create-response *request*)}
  ([err fib]
   server-error-response))
