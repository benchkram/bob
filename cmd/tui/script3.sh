cleanup() {
  echo "redis interrupted"
  exit 0
}

trap cleanup INT

i=0
echo "redis running 0"

while sleep 1
do
  ((i++))
  echo "redis running $i"
done

echo "redis exited"
exit 0
