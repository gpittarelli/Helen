#!/bin/sh

go get github.com/axw/gocov/gocov
go get gopkg.in/matm/v1/gocov-html

gocov test ./controllers/socket/internal/handler | gocov-html > handler.html
gocov test ./models/ | gocov-html > models.html

git clone git@github.com:TF2Stadium/coverage.git
cp handler.html models.html ./coverage/
cd coverage
git config --global user.email "this@is.bot"
git config --global user.name "circleci deploy"
cp index_template index.html
printf "$(date -u) \n</body>" >> index.html
git add models.html handler.html index.html
git commit -m "Update coverage" && git push -f