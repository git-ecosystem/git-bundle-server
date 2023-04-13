@daemon
Feature: Bundle server daemon tests
  Scenario:
    Given the daemon has not been started
    When I run the bundle server CLI command 'web-server start'
    Then the daemon is running

  Scenario:
    Given the daemon was started
    When I run the bundle server CLI command 'web-server stop'
    Then the daemon is not running
