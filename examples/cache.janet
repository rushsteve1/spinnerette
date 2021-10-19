# Example usage of Spinnerette's built-in caching system

# Caches are cleared when Spinnerette is restarted,
# so use SQLite for more permanent data

# Make sure to use a unique enough key since the cache is shared across all
# scripts in a Spinnerette project

# You can interact with the cache using the globally available native functions
# spin/cache-set and spin/cache-get

# The cache-timeout macro lives in the spin module and provides a wrapper around
# the above native functions that also automatically handles timeouts
(import spin)

# Re-runs every 10 minutes (600 seconds)
(spin/cache-timeout :cached-page 600
 (string "I am a cached page that last ran at: " (os/time)))
