#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
if [[ "${TRACE-0}" == "1"  ]]; then
    set -o xtrace
fi

if [[ "${1-}" =~ ^-*h(elp)?$  ]]; then
    echo 'Usage: ./script.sh arg-one arg-two

    -p -s -l

    '
    exit
fi

cd "$(dirname "$0")"

main() {
    echo Header info here
    
    
    for i in "$@"
    do
        case $i in
            -p=*|--prefix=*)
                PREFIX="${i#*=}"
                
            ;;
            -s=*|--searchpath=*)
                SEARCHPATH="${i#*=}"
            ;;
            -l=*|--lib=*)
                DIR="${i#*=}"
            ;;
            --default)
                DEFAULT=YES
            ;;
            *)
                # unknown option
            ;;
        esac
    done
    echo PREFIX = ${PREFIX}
    echo SEARCH PATH = ${SEARCHPATH}
    echo DIRS = ${DIR}
    echo DEFAULT = ${DEFAULT}
    
    case $1 in
        
        Lithuania)
            echo -n "Lithuanian"
        ;;
        
        Romania | Moldova)
            echo -n "Romanian..."
            
        ;;
        
        Italy | "San Marino" | Switzerland | "Vatican City")
            echo -n "Italian"
        ;;
        
        *)
            echo -n "unknown"
        ;;
    esac
    
    
}

main "$@"
