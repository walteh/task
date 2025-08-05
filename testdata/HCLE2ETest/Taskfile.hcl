version = "3"

vars {
  ORIGINAL = "foo"
}

env {
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
    "echo PATH=${env.PATH_COPY}"
  ]
}
