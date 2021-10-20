(defn cache-get
  "(cache-get key &opt dflt)"
  [key &opt dflt]
  (let [[v t] (spinternal/raw-cache-get)]
    (if (and (nil? v) (< t 0))
      (or dflt nil)
      [v t])))

(defn cache-set
  "(cache-set key value)"
  [key value]
  (spinternal/raw-cache-set key value))

(defmacro with-timeout
  "
  (with-timeout key timeout & body)

  Evaluates the given body and stores it in the Spinnerette cache
  with the given `key` (a keyword).
  Subsequent calls will use the cached version.
  Once the `timeout` (integer seconds) has passed the body will be re-run the
  body and cached again.
  "
  [key timeout & body]
  (with-syms [$val $time]
    ~(let [[,$val ,$time] (spinternal/raw-cache-get ,key)]
      # If there is nothing in the cache, or the timeout has passed
      (if (or (nil? ,$val) (> (- (os/time) ,$time) ,timeout))
        (spinternal/raw-cache-set ,key (do ,;body))
        ,$val))))
