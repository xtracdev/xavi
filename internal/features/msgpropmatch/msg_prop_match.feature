Feature: Message Property Matching

  @msgpropmatch
  Scenario:
    Given Routes with msgprop expressions
    And The routes have a common uri
    Then Requests are dispatched based on msgprop matching