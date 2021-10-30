# Spinnerette bundles in the janet-html library for easily creating HTML pages
# with pure Janet. It uses a syntax similar to Clojure's Hiccup

(import janet-html :as html)

(html/encode
 [:html
  [:body
   [:h1 "Hello from Janet-HTML"]
   [:p "this was created with pure Janet!"]]])
