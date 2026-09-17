# Agent-Engineering Java V2.5.8
# 1:1 port of agent-engineering (Python) / agent-engineering-go (Go).

JAR = target/agent-engineering-java-2.5.8.jar

.PHONY: build run test verify tenant pipeline trace clean

build:
	mvn compile

test:
	mvn test

$(JAR):
	mvn package -DskipTests

run: $(JAR)
	java -jar $(JAR)

verify: $(JAR)
	java -jar $(JAR) verify

tenant: $(JAR)
	java -jar $(JAR) tenant

pipeline: $(JAR)
	java -jar $(JAR) pipeline

trace: $(JAR)
	java -jar $(JAR) trace

clean:
	mvn clean
	rm -rf target
