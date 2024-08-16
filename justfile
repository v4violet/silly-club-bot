
set dotenv-load


out := './dist/' + if os_family() == 'windows' { 'program.exe' } else { 'program' }

static := env_var_or_default("BUILD_STATIC", "false")

modules := env_var_or_default("BUILD_MODULES", "all")

tags := replace(prepend("modules.", replace(modules, ",", " ")), " ", ",")

ldflags := if static == "true" { "-w -s" } else { "" }


build_flags_ldflags := "-ldflags " + quote(ldflags)

build_flags_tags := if tags == "" { "" } else { "-tags=" + quote(tags) }

build_flags_out := "-o " + out

build_flags_test := trim(build_flags_ldflags + " " + build_flags_tags)

build_flags := build_flags_test + " " + build_flags_out


default:
    @just --list
    
tidy:
    go mod tidy

test:
    go test {{build_flags_test}}

build:
    go build {{build_flags}}

run: build
    {{out}}

docker-build:
    docker build --build-arg="BUILD_MODULES={{modules}}" -t silly-club-bot .

docker-run: docker-build
    docker run --rm --env-file .env silly-club-bot