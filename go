#!/bin/bash

CMD=${0##*/}
OPS="app-show,app-ssh,app-restart,app-stop,app-scale-down,app-show --configuration,app-show --state,app-show --gears,customized command"
DATA_DIR=$HOME/data
OVERWRITE=${OVERWRITE:-1}

usage() {
    cat <<EOU
usage: $CMD [-b] [-g]
options:
    -b
    Batch mode. Execute commands for list of domain/app.

    -g
    Genesis mode. Collect domain/app data in local data cache.
EOU
}

err() {
    msg="$@"
    echo "ERROR: $msg" >&2
}


generate_data() {
    echo listing authorization
    (
        # fix stdin in case of piped stdin
        exec < /dev/tty
        # precheck
        rhc authorization list < /dev/null &> /dev/null || {
            rhc account || exit 1
            [[ $batch_mode -eq 1 ]] && exit 1
        }
    )
    t0=$(date +%s)
    dc=0
    ac=0
    if [[ $batch_mode -eq 1 ]]
    then
        domains=$(< /dev/stdin)
    else
        domains=$(rhc domain list | grep Domain| awk '{print $2}'| sort -u)
        select dom in ALL $domains
        do
            [[ -n $dom ]] && break
        done
        [[ $dom == "ALL" ]] || domains=$dom
    fi
    for dom in $domains
    do
        dom=${dom#data/}
        echo "processing domain '$dom' ..."
        for app in $(rhc domain show -n $dom | grep uuid | awk '{print $1}')
        do
            # In UPDATE mode, bypass existing ones
            if [[ $gendata_update -eq 1 ]]
            then
                test -d $DATA_DIR/$dom/$app && continue
            fi
            echo "processing domain '$dom' application '$app' ..."
            mkdir -p $DATA_DIR/$dom/$app
            _rc=0
            [[ -e $DATA_DIR/$dom/$app/show.txt && $OVERWRITE -eq 0 ]] || {
                rhc app show -n $dom -a $app > $DATA_DIR/$dom/$app/show.txt || {
                    err failed to show $dom/$app
                    _rc=1
                }
            }
            [[ -e $DATA_DIR/$dom/$app/show_gears.txt && $OVERWRITE -eq 0 ]] || {
                rhc app show --gears -n $dom -a $app > $DATA_DIR/$dom/$app/show_gears.txt || {
                    err failed to show gears $dom/$app
                    _rc=1
                }
            }
            [[ -e $DATA_DIR/$dom/$app/show_state.txt && $OVERWRITE -eq 0 ]] || {
                rhc app show --state -n $dom -a $app > $DATA_DIR/$dom/$app/show_state.txt || {
                    err failed to show state $dom/$app
                    _rc=1
                }
            }
            [[ $_rc -eq 0 ]] || {
                err "failed in processing $dom/$app"
            }
            ((ac++))
        done
        ((dc++))
    done
    t1=$(date +%s)
    (( dt = t1 - t0 ))
    echo Time elapsed: $dt seconds, $dc domains and $ac applications
}

do_cmd() {
    cmd=$@

    [[ $batch_mode -eq 0 || -n "$RHCW_SHOW_CMD" ]] && echo "CMD: $cmd"
    if [[ $INTERACTIVE -eq 1 ]]
    then
        read -p 'Are you sure to continue[n]?' a
        [[ ${a:0:1} == 'y' || ${a:0:1} == 'Y' ]] || exit 1
    fi

    # fix stdin in case of piped stdin
    exec < /dev/tty
    $cmd
    rc=$?
    [[ $batch_mode -eq 1 ]] || {
        echo
        cat <<EOC
CMD: (copied to clipboard and bash history)
        $cmd

EOC
        echo "$cmd" | xclip 2>/dev/null
        (sleep 1; echo "$cmd" >> ~/.bash_history ) &
    }
}

# --------
#   MAIN
# --------

batch_mode=${BATCH_MODE:-0}
lae_data=''
gendata=0
gendata_update=0
while getopts ":bd:ghu" opt
do
    case $opt in
        b) batch_mode=1 ;;
        d) [[ -f $OPTARG ]] && lae_data=$(< $OPTARG) ;;
        g) gen_data=1 ;;
        h) usage
           exit 0
        ;;
        u) gendata_update=1 ;;
    esac
done

if [[ $gen_data -eq 1 ]]
then
    generate_data
    exit 0
fi

(
# fix stdin in case of piped stdin
exec < /dev/tty
# precheck
rhc authorization list < /dev/null &> /dev/null || {
    rhc account || exit 1
    [[ $batch_mode -eq 1 ]] && exit 1
}
)


[[ $OPTIND -gt 0 ]] && shift $((OPTIND-1))

if [[ $batch_mode -eq 1 ]]
then
    [[ -n "$lae_data" ]] || {
        lae_data=$(< /dev/stdin)
    }

    [[ -n "$lae_data" ]] || {
        err "LAE data is required in batch mode"
        err "Provide LAE data at STDIN or '-d <lae_data_file>'"
        exit 1
    }

    op=${@:-'app-show'}

    for ent in $lae_data; do
        ent=${ent#data/}
        app=${ent#*/}
        dom=${ent%/*}
        cmd="rhc $op -a $app -n $dom"
        do_cmd "$cmd"
    done

else

    cat <<EOM
    IMPORTANT: Be sure to run '$CMD -g'(to collect domain/app data) at least once before using it
EOM
    PS3='>> Select operation ? '
    OLD_IFS=$IFS; IFS=,
    select op in $OPS; do
        [[ -n "$op" ]] && break;
    done
    IFS=$OLD_IFS

    PS3=">> rhc $op -n <domain> ? "
    #COLUMNS=1
    select dom in $(cd $DATA_DIR; ls -1)
    do
        [[ -n "$dom" ]] && break;
    done

    PS3=">> rhc $op -n $dom -a <application> ?"
    expr $dom : '.*nprd' >/dev/null || PS3=">> rhc $op -n $dom -a <application> ?(*prd1:PRD1,*3.prd:DR,*4.prd:PROD)"

    COLUMNS=1
    select app in $(cd $DATA_DIR/$dom; ls -1)
    do
        [[ -n "$app" ]] && break;
    done
    unset COLUMNS

    if [[ ${op:0:10} == 'customized' ]]; then
        read -p 'command: ' op
        cmd="$op -a $app -n $dom"
    else
        cmd="rhc $op -a $app -n $dom"
    fi

    do_cmd $cmd

fi
exit $rc
