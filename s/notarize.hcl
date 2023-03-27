apple_id {
  username = "@env:AC_USERNAME"
  password = "@env:AC_PASSWORD"
}
notarize {
  path = "observe.dmg"
  bundle_id = "com.observeinc.developer.observe"
  staple = true
}
notarize {
  path = "observe.zip"
  bundle_id = "com.observeinc.developer.observe"
  staple = true
}
notarize {
  path = "observe.pkg"
  bundle_id = "com.observeinc.developer.observe"
  staple = true
}
