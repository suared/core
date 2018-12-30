#I am going to use dev to represent both local and dev (Shared) to start, will split when makes sense...

#Things to keep in mind…  Reads are eventually consistent (hard to tell on local computer)
#For a complete list of DynamoDB runtime options, including -port , type this command:
#java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -help
#CLI example
# aws dynamodb list-tables --endpoint-url http://localhost:8000

#Run from Ubuntu w/ Docker, note: VBox port forwarding setup for this port
docker run -p 9001:9001 -it --rm amazon/dynamodb-local -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -port 9001 -inMemory -delayTransientStatuses