version = "3"

vars {
  ORIGINAL = "foo"
  NAME = "BOB"
  GREETING = "Hello, ${vars.NAME}!"
  UPPER_GREETING = upper(vars.GREETING)
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

task "all" {
  deps = [
    task("build"),
    task("lint", {MODE = "fast"})
  ]
  cmds = [
    "echo FINAL ${vars.ORIGINAL}",
    "echo PATH=${env.PATH_COPY}",
    "echo GREET=${vars.UPPER_GREETING}",
    "echo EXT=${env.EXTENDED}"
  ]
}
