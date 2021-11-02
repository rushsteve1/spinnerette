(import spork/temple)
(import janet-html :as html)
(temple/add-loader)
(import ./base :as base)

(def header [:h1 "Spinnerette"])
(def main [:p "The Spinnerette website"])

(let [out @""]
    (with-dyns [:out out]
        (base/render :header (html/encode header) :main (html/encode main)))
    out)