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
  vars { MODE = "fast" }
  cmds = ["echo LINT MODE ${vars.MODE}", "echo TWO-DONE"]
}

task "all" {
  deps = [
    task.build,
    task.lint,
    task.scoped,
  ]
  cmds = [
    "echo FINAL ${vars.ORIGINAL}",
    "echo PATH=${env.PATH_COPY}",
    "echo GREET=${vars.UPPER_GREETING}",
    "echo EXT=${env.EXTENDED}",
    "echo PLATFORM=${vars.SUPPORTED_PLATFORMS[0]}"
  ]
}

task "test" {
  cmds = ["echo 'echo TEST'"]
}

task "delayed" {
  vars { MSG = "" }
  cmds = ["echo 'echo DELAYED ${vars.MSG}'"]
}

task "scoped" {
  deps = [
    task.test,
    task.delayed,
  ]
  cmds = [
    exec(task.test),
    exec(task.delayed, { MSG = "DELAYED" }),
  ]
}
