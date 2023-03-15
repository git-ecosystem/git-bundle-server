Feature: Basic bundle server usage

  Background: The bundle web server is running
    Given the bundle web server was started at port 8080

  Scenario: A user can clone with a bundle URI pointing to the bundle server
    Given a remote repository 'https://github.com/vdye/asset-hash.git'
    Given the bundle server has been initialized with the remote repo
    When I clone from the remote repo with a bundle URI
    Then bundles are downloaded and used
