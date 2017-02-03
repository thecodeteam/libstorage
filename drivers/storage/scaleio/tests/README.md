# ScaleIO driver testing

ScaleIO driver requires valid AWS environment in which tests can run. The
requirements are:

* **VPC**: Private network where EC2 instance and ScaleIO instance can be launched

* **Subnet**: In VPC it requires at least one valid Subnet for EC2 instance and
  ScaleIO

* **EC2**: Any EC2 instance that has permission to run ScaleIO plugin. For list of
  IAM permissions see user documentation.

For automated testing there are couple script available that will spin up
while AWS environment by using CloudFormation service.

* `./test-env-up.sh [access-key] [secret-key] [stack-name] [key-name]` -
  should be used to launch whole AWS environment required for ScaleIO storage
  driver testing.
  `[access-key]: AWS access key`
  `[secret-key]: AWS secret key`
  `[stack-name]: AWS stack name that must be uniquely identifiable`
  `[key-name]: AWS key that will be used to launch EC2 instance`

* `./test-run.sh [stack-name] [rexray-path]` - to run the tests on EC2 instance.
  `[stack-name]` is the same name used in the setup process.
  `[rexray-path]` is the fully qualified path to the REX-Ray binary to test.

* `./test-env-down.sh [stack-name]` - to tear down AWS environment infrastructure
  and clean up all resources.
  `[stack-name]` is the same name used in the setup and testing process.
