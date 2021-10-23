(defn cache-get
  "Get a value from the cache with an optional default."
  [key &opt dflt]
  (get *cache* key dflt))

(defn cache-set
  "Set a value in the cache"
  [key value]
  (put *cache* key [value (os/time)]))

(defn cache-del
  "Delete a value from the cache. Same as setting nil."
  [key]
  (put *cache* key nil))

(defmacro with-timeout
  "
  Evaluates the given body and stores it in the Spinnerette cache
  with the given `key` (a keyword).
  Subsequent calls will use the cached version.
  Once the `timeout` (integer seconds) has passed the body will be re-run the
  body and cached again.
  "
  [key timeout & body]
  (with-syms [$val $time]
    ~(let [[,$val ,$time] (,cache-get ,key [nil -1])]
      # If there is nothing in the cache, or the timeout has passed
      (if (or (nil? ,$val) (> (- (os/time) ,$time) ,timeout))
        (let [ret (do ,;body)]
          (,cache-set ,key ret)
          ret)
        ,$val))))
