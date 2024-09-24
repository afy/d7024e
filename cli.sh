echo "$@" > /tmp/kademlia_pipe
read -r line < /tmp/kademlia_resp
echo "$line"
