Feature: Auth configuration on the web server

  Background: The bundle server has an initialized route
    Given no bundle server repository exists at route 'integration/auth'
    Given a new remote repository with main branch 'main'
    Given the remote is cloned
    Given 5 commits are pushed to the remote branch 'main'
    Given a bundle server repository is created at route 'integration/auth' for the remote

  Scenario: With no auth config, bundle list can be accessed anonymously
    Given the bundle web server was started at port 8080
    When I request the bundle list
    Then the response code is 200
    Then the response is a valid bundle list

  Scenario: If basic auth is required but none is sent, get a 401
    Given an auth config with username 'my_username' and password 'p4sSW0rD!'
    Given the bundle web server was started at port 8080 with auth config
    When I request the bundle list
    Then the response code is 401
    Then the response is empty

  Scenario: If basic auth is required and we send bad credentials, get a 404
    Given an auth config with username 'my_username' and password 'p4sSW0rD!'
    Given the bundle web server was started at port 8080 with auth config
    When I request the bundle list with username 'my_userName' and password 'password!'
    Then the response code is 404
    Then the response is empty

  Scenario: If basic auth is required and we send the right credentials, can access the bundle list
    Given an auth config with username 'my_username' and password 'p4sSW0rD!'
    Given the bundle web server was started at port 8080 with auth config
    When I request the bundle list with username 'my_username' and password 'p4sSW0rD!'
    Then the response code is 200
    Then the response is a valid bundle list
