package testsupport

//roundRobin3000Config is the mountebank imposter definition for the test server running on port 3000
//that xavi proxies
const RoundRobin3000Config = `
{
  "port": 3000,
  "protocol": "http",
  "stubs": [
    {
      "responses": [
        {
          "is": {
            "statusCode": 200,
            "body": "All work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\n"
          }
        }
      ],
      "predicates": [
        {
          "equals": {
            "path": "/hello",
            "method": "GET"
          }
        }
      ]
    }
  ]
}

`

//roundRobin3100Config is the mountebank imposter definition for the test server running on port 3000
//that xavi proxies
const RoundRobin3100Config = `
{
  "port": 3100,
  "protocol": "http",
  "stubs": [
    {
      "responses": [
        {
          "is": {
            "statusCode": 200,
            "body": "All work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\n"
          }
        }
      ],
      "predicates": [
        {
          "equals": {
            "path": "/hello",
            "method": "GET"
          }
        }
      ]
    }
  ]
}
`
