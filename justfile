
set dotenv-load


package_name := 'github.com/v4violet/silly-club-bot'

out := './dist/' + if os_family() == 'windows' { 'program.exe' } else { 'program' }

static := env_var_or_default("BUILD_STATIC", "false")

modules := env_var_or_default("BUILD_MODULES", "all")

tags := replace(prepend("modules.", replace(modules, ",", " ")), " ", ",")

git_pending_changes := trim(shell('git status --porcelain | wc -l'))


ldflag_build_pkg := package_name + '/build'

ldflag_version_suffix := if git_pending_changes == "0" { "+" + trim(shell('git rev-parse --short HEAD')) } else { "-dev" }

ldflag_version := '-X ' + ldflag_build_pkg + '.Version=' + datetime_utc('%Y.%m.%d') + ldflag_version_suffix

ldflag_static := if static == "true" { "-w -s" } else { "" }

ldflags := trim(ldflag_version + ' ' + ldflag_static)


build_flags_ldflags := "-ldflags " + quote(ldflags)

build_flags_tags := if tags == "" { "" } else { "-tags=" + quote(tags) }

build_flags_out := "-o " + out

build_flags := trim(build_flags_ldflags + " " + build_flags_tags) + " " + build_flags_out


default:
    @just --list
    
tidy:
    go mod tidy

build:
    go build {{build_flags}}

run: build
    {{out}}

docker-build:
    docker build --build-arg="BUILD_MODULES={{modules}}" -t silly-club-bot .

docker-run: docker-build
    docker run --rm --env-file .env silly-club-bot