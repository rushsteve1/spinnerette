# This file will be evaluated when the Janet VM is started by Spinnerette
# BEFORE any modules are loaded but AFTER spinternal is loaded

(def bundled-paths :private
  [ "libs/spork/spork/:all:.janet"
    "libs/spin/:all:.janet"
    "libs/spork/:all:/init.janet"
    "libs/janet-html/src/:all:.janet"
    "libs/:all:/init.janet"
    "libs/spork/:all:.janet"
    "libs/:all:.janet" ])

(defn- is-bundled [path]
  (var ret nil)
  (each p bundled-paths
    (let [fullpath (string (module/expand-path path p))]
      (when (in module/cache fullpath)
        (set ret fullpath)
        (break))))
  ret)

(array/push module/paths [is-bundled :preload])