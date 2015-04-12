# Copyright 2014 The Cactus Authors. All rights reserved.

GOFILES = api/account.go api/activity.go api/clarification.go api/contest.go api/execution.go api/notification.go api/problem.go api/session.go api/standing.go api/submission.go belt/belt.go belt/blobs.go belt/execution.go belt/queue-local.go belt/queue-remote.go belt/queue.go belt/stack-c.go belt/stack-cpp.go belt/stack.go belt/submission.go cmd/cactus/main.go cmd/cactus/router.go data/account.go data/blobs.go data/clarification.go data/contest.go data/db.go data/execution.go data/notification.go data/problem.go data/standing.go data/submission.go hub/hub.go cube/proc.go cube/cube.go cube/docker.go cube/errors.go cube/plain.go cube/proc.go rpc/blobs.go rpc/executions.go rpc/problems.go rpc/queue.go rpc/submissions.go ui/assets.go ui/pages.go

CSSFILES = ui/assets/css/animate+animo.css ui/assets/css/bootstrap.css ui/assets/css/fontawesome.css ui/assets/css/hightlight.css ui/assets/css/nprogress.css ui/assets/css/screen.css

JSFILES = ui/assets/js/underscore.js ui/assets/js/jquery.js ui/assets/js/backbone.js ui/assets/js/bootstrap.js ui/assets/js/bootbox.js ui/assets/js/sugar.js ui/assets/js/moment.js ui/assets/js/nprogress.js ui/assets/js/animo.js ui/assets/js/async.js ui/assets/js/lunr.js ui/assets/js/showdown.js ui/assets/js/showdown-github.js ui/assets/js/showdown-table.js ui/assets/js/highlight.js ui/assets/js/cactus.js

RSCFILES = cmd/cactus/config-sample.tml data/db-init.sql ui/assets/css/screen.min.css ui/assets/font/fontawesome.eot ui/assets/font/fontawesome.svg ui/assets/font/fontawesome.ttf ui/assets/font/fontawesome.woff ui/assets/font/glyphicons-halflings-regular.eot ui/assets/font/glyphicons-halflings-regular.svg ui/assets/font/glyphicons-halflings-regular.ttf ui/assets/font/glyphicons-halflings-regular.woff ui/assets/font/mavenpro-bold.eot ui/assets/font/mavenpro-bold.svg ui/assets/font/mavenpro-bold.ttf ui/assets/font/mavenpro-bold.woff ui/assets/font/mavenpro-regular.eot ui/assets/font/mavenpro-regular.svg ui/assets/font/mavenpro-regular.ttf ui/assets/font/mavenpro-regular.woff ui/assets/font/ubuntumono-r.eot ui/assets/font/ubuntumono-r.svg ui/assets/font/ubuntumono-r.ttf ui/assets/font/ubuntumono-r.woff ui/assets/img/mascot.png ui/assets/js/cactus.min.js ui/assets/json/credits.json ui/index.min.html cubed/cubed.go

all: cactus

clean:
	rm -f cactus.zip cactus ui/index.min.html ui/assets/css/screen.min.css ui/assets/js/cactus.min.js.map ui/assets/js/cactus.min.js

cactus: $(GOFILES) $(RSCFILES)
	go build ./cmd/cactus
	zip - $(RSCFILES) | cat >> $@
	zip -A $@

ui/assets/css/screen.min.css: $(CSSFILES)
	cat $^ | cleancss --s0 --s1 -o $@

ui/assets/js/cactus.min.js: $(JSFILES)
	uglifyjs $^ -c -m --screw-ie8 -p 1 --source-map $@.map --source-map-include-sources --source-map-root / --source-map-url /assets/js/cactus.min.js.map > $@

ui/index.min.html: ui/index.html
	cat $^ | tr '\t\n' '  ' | sed -e 's/  */ /g' > $@

cactus.zip: cactus LICENSE
	zip $@ $^
