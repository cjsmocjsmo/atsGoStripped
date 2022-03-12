# for f in /home/charliepi/go/atsGo/static/gallery/landscape/*.jpg; do
# cwebp -q 95 -resize 300 0 "$f" -o "${f%.jpg}_thumb.webp"
# done

for f in /home/charliepi/go/atsGo/static/images/*.jpg; do
cwebp -q 95 "$f" -o "${f%.jpg}.webp"
done
