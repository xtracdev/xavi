
Feature: Prefer local routing
  This tests a backend configured with a prefer-local load balancer

  @preflocal
  Scenario: Basic prefer local
    Given A preflocal route with backend definitions with two servers
    And The load balancing policy is prefer local
    And I send two requests to the prelocal listener
    Then Only the local server handles the requests