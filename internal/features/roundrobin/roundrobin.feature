
Feature: Round robin routing
  This tests a backend configured with a round robin load balancer

  @roundrobin @basicroundrobin
  Scenario: Basic round robin
    Given I have a backend definitions with two servers
    And The load balancing policy is round robin
    And I send two requests to the listener
    Then Each server gets a single request