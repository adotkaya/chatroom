build-chat:
	@go build -o ./bin/chat .

chat: build-chat
	@./bin/chat


test-chat-race:
	@go clean -testcache
	@go test -v ./...


echo "# chatroom" >> README.md
git init
git add README.md
git commit -m "git push commit"
git branch -M master
git remote add origin https://github.com/adotkaya/chatroom.git
git push -u origin master
