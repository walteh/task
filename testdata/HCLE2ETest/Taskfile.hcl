version = "3"

vars {
  ORIGINAL = "foo"
  NAME = "BOB"
  GREETING = "Hello, ${vars.NAME}!"
  UPPER_GREETING = upper(vars.GREETING)
  SUPPORTED_PLATFORMS = [
    "linux", "darwin", "windows",
  ]
  DOCKER_OPTIONS = {
    cache = true
    platform = vars.SUPPORTED_PLATFORMS[0]
  }
  ABC = 123
}

env {
  EXTENDED = "${env.BASE}-ext"
  BASE = "base"
  PATH_COPY = env("PATH")
}

task "build" {
  vars { VERSION = sh("echo 1.2.3") }
  cmds = ["echo BUILD:${vars.VERSION}", "echo ONE-DONE"]
}

task "lint" {
  cmds = ["echo LINT MODE ${vars.MODE}", "echo TWO-DONE"]
}

task "scoped" {
  vars {
    DEF = 456
    GHI = "${vars.ABC}${vars.DEF}"
  }
  cmds = ["echo SCOPED ${vars.ABC} ${vars.DEF} ${vars.GHI}"]
}

task "all" {
  deps = [
    task("build"),
    task("lint", {MODE = "fast"}),
    task("scoped"),
  ]
  cmds = [
    "echo FINAL ${vars.ORIGINAL}",
    "echo PATH=${env.PATH_COPY}",
    "echo GREET=${vars.UPPER_GREETING}",
    "echo EXT=${env.EXTENDED}",
    "echo PLATFORM=${vars.SUPPORTED_PLATFORMS[0]}"
  ]
}
