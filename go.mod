module github.com/omec-project/gnbsim

go 1.15

require (
	git.cs.nctu.edu.tw/calee/sctp v1.1.0
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/aws/aws-sdk-go v1.36.24 // indirect
	github.com/calee0219/fatal v0.0.1
	github.com/free5gc/CommonConsumerTestData v1.0.0
	github.com/free5gc/MongoDBLibrary v1.0.0
	github.com/free5gc/UeauCommon v1.0.0
	github.com/free5gc/amf v1.3.0
	github.com/free5gc/aper v1.0.0
	github.com/free5gc/idgenerator v1.0.0
	github.com/free5gc/logger_util v1.0.0
	github.com/free5gc/milenage v1.0.0
	github.com/free5gc/ngap v1.0.1
	github.com/free5gc/openapi v1.0.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/omec-project/nas v1.0.6-gnbsim
	github.com/onosproject/onos-api/go v0.9.15
	github.com/onosproject/onos-lib-go v0.8.16
	github.com/sirupsen/logrus v1.8.1
	github.com/ugorji/go v1.2.3 // indirect
	github.com/urfave/cli v1.22.4
	github.com/yerden/go-util v1.1.4
	go.mongodb.org/mongo-driver v1.4.4
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e
	google.golang.org/grpc v1.41.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200229013735-71373c6105e3
