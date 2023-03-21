Feature: Bundle server command tests

  Background: The bundle web server is running
    Given the bundle web server was started at port 8080

  @online
  Scenario: The init command initializes a bundle server repository
    Given no bundle server repository exists at route 'integration/asset-hash'
    When I run the bundle server CLI command 'init https://github.com/vdye/asset-hash.git integration/asset-hash'
    Then a bundle server repository exists at route 'integration/asset-hash'

  @online
  Scenario: The delete command removes route configuration and repository data
    Given a remote repository 'https://github.com/vdye/asset-hash.git'
    Given a bundle server repository is created at route 'integration/asset-hash' for the remote
    When I run the bundle server CLI command 'delete integration/asset-hash'
    Then the route configuration and repository data at 'integration/asset-hash' are removed

  Scenario: The update command fetches the latest remote content and updates the bundle list
    Given no bundle server repository exists at route 'integration/bundle'
    Given a new remote repository with main branch 'main'
    Given the remote is cloned
    Given 5 commits are pushed to the remote branch 'main'
    Given a bundle server repository is created at route 'integration/bundle' for the remote
    Given 2 commits are pushed to the remote branch 'main'
    When I run the bundle server CLI command 'update integration/bundle'
    Then the bundles are fetched and the bundle list is updated

  Scenario: The stop command updates the routes file
    Given no bundle server repository exists at route 'integration/stop'
    Given a new remote repository with main branch 'main'
    Given a bundle server repository is created at route 'integration/stop' for the remote
    When I run the bundle server CLI command 'stop integration/stop'
    Then the route is removed from the routes file

  Scenario: The start command updates the routes file
    Given no bundle server repository exists at route 'integration/start'
    Given a new remote repository with main branch 'main'
    Given a bundle server repository is created at route 'integration/start' for the remote
    When I run the bundle server CLI command 'stop integration/start'
    When I run the bundle server CLI command 'start integration/start'
    Then the route exists in the routes file
