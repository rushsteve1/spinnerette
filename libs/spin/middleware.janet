# These middleware are taken from Circlet
# https://github.com/janet-lang/circlet/blob/master/circlet_lib.janet

# Copyright (c) 2018 Calvin Rose
# Copyright (c) 2004-2013 Sergey Lyubka <valenok@gmail.com>
# Copyright (c) 2013-2018 Cesanta Software Limited

(defn middleware
  "Coerce any type to http middleware"
  [x]
  (case (type x)
    :function x
    (fn [&] x)))

(defn router
  "Creates a router middleware. Route parameter must be table or struct
  where keys are URI paths and values are handler functions for given URI"
  [routes]
  (fn [req]
    (def r (or
             (get routes (get req :uri))
             (get routes :default)))
    (if r ((middleware r) req) 404)))

(defn logger
  "Creates a logging middleware. nextmw parameter is the handler function
  of the next middleware"
  [nextmw]
  (fn [req]
    (def {:uri uri
          :protocol proto
          :method method
          :query-string qs} req)
    (def start-clock (os/clock))
    (def ret (nextmw req))
    (def end-clock (os/clock))
    (def fulluri (if (< 0 (length qs)) (string uri "?" qs) uri))
    (def elapsed (string/format "%.3f" (* 1000 (- end-clock start-clock))))
    (def status (or (get ret :status) 200))
    (print proto " " method " " status " " fulluri " elapsed " elapsed "ms")
    ret))

(defn cookies
  "Parses cookies into the table under :cookies key. nextmw parameter is
  the handler function of the next middleware"
  [nextmw]
  (def grammar
    (peg/compile
      {:content '(some (if-not (set "=;") 1))
       :eql "="
       :sep '(between 1 2 (set "; "))
       :main '(some (* (<- :content) :eql (<- :content) (? :sep)))}))
  (fn [req]
    (-> req
        (put :cookies
             (or (-?>> [:headers "Cookie"]
                       (get-in req)
                       (peg/match grammar)
                       (apply table))
                 {}))
        nextmw)))