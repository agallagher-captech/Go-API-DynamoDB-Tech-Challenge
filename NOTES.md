# Masitda Monday

## June 5 2025
Input model validation - added decodeValid interface, will use for put/post requests where the request has a body
Started Create User functionality, struggled with whether to have `routes.go` have just the url or "POST /api/users" (copilot says does not http serve mux does not support having POST at the beginning), need to parse through this error:
```
"level":"ERROR","msg":"failed to create user","error":"[in main.UsersService.CreateUser] failed to put item: operation error DynamoDB: PutItem, https response error StatusCode: 400, RequestID: e5978ea9-e877-454a-85cc-d5282cfa5754, api error ValidationException: One or more parameter values are not valid. A value specified for a secondary index key is not supported. The AttributeValue for a key attribute cannot contain an empty string value. IndexName: GSI1, IndexKey: GSI1SK"}
```
## Adding Health Check and Setting Up Unit Tests with Mockery

### Summary of Changes
Today, I made significant progress in the project by implementing a health check endpoint, writing my first unit test, and setting up a framework for mocking dependencies using `mockery`. Below are the key updates:

1. **Health Check Implementation**:
   - Added a new health check handler in `handlers/health.go`. This endpoint provides a simple way to verify the application's availability and responsiveness.
   - Wrote the first unit test for the project in `handlers/health_test.go`, ensuring the health check handler behaves as expected.

2. **Unit Test for User Service**:
   - Started writing a unit test for the `ReadUser` method in `services/user_test.go`. This required mocking the DynamoDB client to simulate database interactions without making actual AWS calls.

3. **Setting Up Mockery**:
   - Configured a `.mockery.yaml` file to streamline the process of generating mocks for interfaces. This file specifies:
     - The output directory for generated mocks (`mocks`).
     - Naming conventions for mock types (e.g., camel case).
   - Used the `make mock-gen` command to generate a mock for the DynamoDB client. This mock is now used in the `user_test.go` file to test the `ReadUser` method.
     - Updated it to use the tag `@mockery --all` to resolve this error: ```EDT FTL Use --name to specify the name of the interface or --all for all interfaces found dry-run=false version=v2.49.0```

### Insights on Mockery
Mockery is a tool that automates the creation of mock implementations for interfaces, making it easier to write unit tests for code that depends on external systems or services. Here’s how it works in this project:
- **Interface-Based Mocking**: Mockery generates mocks based on the interfaces defined in the codebase. For example, the DynamoDB client interface was used to generate a mock that simulates its behavior.
- **Configuration with `mockery.yaml`**: The `mockery.yaml` file ensures consistency in how mocks are generated. It defines the output directory (`mocks`) and enforces a naming convention (camel case) for the generated mock types.
- **Integration with Testify**: The generated mocks integrate seamlessly with the `testify` library, allowing for easy setup of expectations and assertions in unit tests.

By using `mockery`, I was able to focus on writing meaningful tests without the overhead of manually creating mock implementations. This setup ensures that the tests are maintainable and aligned with the project's structure.

### Next Steps
- Complete the unit test for the `ReadUser` method in `services/user_test.go`.
- Add more unit tests for other services and handlers to improve test coverage.
- Add Create User functionality.

Today’s work established a strong foundation for testing and mocking in the project, ensuring that future development is well-supported by robust testing practices.
