cleanup() {
  echo "mongo interrupted"
  exit 0
}

trap cleanup INT

i=0
echo "mongo running 0"

while sleep .5
do
  ((i++))
  echo "mongo running $i"
done

echo "mongo exited"
exit 0
