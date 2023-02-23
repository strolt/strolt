architecture=""

platform="$(uname | tr '[:upper:]' '[:lower:]')"

if [ "$platform" == "darwin" ]; then
    case $(uname -m) in
        x86_64) architecture="amd64" ;;
        arm64) architecture="arm64" ;;
    esac
else
    case $(uname -m) in
        i386)   architecture="386" ;;
        i686)   architecture="386" ;;
        x86_64) architecture="amd64" ;;
        arm)    dpkg --print-architecture | grep -q "arm64" && architecture="arm64" || architecture="arm" ;;
    esac
fi

echo ""$platform"_"$architecture""
