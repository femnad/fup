from golang:bookworm
run mkdir -p /root/fup
copy . /root/fup/
workdir /root/fup
run go install
entrypoint ["fup"]
