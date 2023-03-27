apple_id {
  username = "@env:AC_USERNAME"
  password = "@env:AC_PASSWORD"
}
notarize {
  path = "dist/observe-darwin_darwin_arm64/observe"
  bundle_id = "com.observeinc.developer.observe"
  staple = true
}
notarize {
  path = "dist/observe-darwin_darwin_amd64/observe"
  bundle_id = "com.observeinc.developer.observe"
  staple = true
}
