Feature: Healthcheck

  @healthcheck @withunhealthy
  Scenario: HTTP GET Healthcheck
      Given A backend with some unhealthy servers
      And I invoke a service against the backend
      Then The service calls succeed against the healthy backends

  @healthcheck @withhealed
  Scenario: Healed server
     Given A previously unhealthy server becomes healthy
     And I invoke a service against the backend
     Then The healed backend recieves traffic
