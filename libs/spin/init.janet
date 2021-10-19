(import json :as js)

(defn json
  "
  (spin/json x)

  Create a JSON response
  Takes serializable content like a map, array, string, or number
  Returns a Circlet response map with the appropriate headers with the content
  encoded as a JSON string.
  "
  [content &opt &keys {:status status}]
  {:headers {"Content-Type" "application/json"}
   :status (or status 200)
   :body (js/encode content)})

(defmacro cache-timeout
  "
  (spin/cache-timeout key timeout & body)

  Evaluates the given body and stores it in the Spinnerette cache
  with the given `key` (a keyword).
  Subsequent calls will use the cached version.
  Once the `timeout` (integer seconds) has passed the body will be re-run the
  body and cached again.
  "
  [key timeout & body]
  (with-syms [$val $time]
    ~(let [[,$val ,$time] (spin/cache-get ,key)]
      # If there is nothing in the cache, or the timeout has passed
      (if (or (nil? ,$val) (> (- (os/time) ,$time) ,timeout))
        (spin/cache-set ,key (do ,;body))
        ,$val))))
