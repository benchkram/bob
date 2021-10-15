cleanup() {
  echo "app interrupted"
  exit 0
}

trap cleanup INT

i=0
echo "app running 0"

while sleep .25
do
  ((i++))
  echo "app running $i"
done

echo "app exited"
exit 0
