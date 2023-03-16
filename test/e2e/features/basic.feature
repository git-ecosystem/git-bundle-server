Feature: Basic bundle server usage

  Background: The bundle web server is running
    Given the bundle web server was started at port 8080

  Scenario: A user can clone with a bundle URI pointing to the bundle server
    Given a remote repository 'https://github.com/vdye/asset-hash.git'
    Given the bundle server has been initialized with the remote repo
    When I clone from the remote repo with a bundle URI
    Then bundles are downloaded and used

  Scenario: A user can fetch with a bundle server that's behind and get all updates
    Given a new remote repository with main branch 'main'
    Given another user pushed 10 commits to 'main'
    Given the bundle server has been initialized with the remote repo
    Given I cloned from the remote repo with a bundle URI
    Given another user pushed 2 commits to 'main'
    When I fetch from the remote
    Then I am up-to-date with 'main'
    Then my repo's bundles are not up-to-date with 'main'

  Scenario: A user will fetch incremental bundles to stay up-to-date
    Given a new remote repository with main branch 'main'
    Given another user pushed 10 commits to 'main'
    Given the bundle server has been initialized with the remote repo
    Given I cloned from the remote repo with a bundle URI
    Given another user pushed 2 commits to 'main'
    Given the bundle server was updated for the remote repo
    When I fetch from the remote
    Then I am up-to-date with 'main'
    Then my repo's bundles are up-to-date with 'main'

  Scenario: A user can fetch force-pushed refs from the bundle server
    Given a new remote repository with main branch 'main'
    Given another user pushed 10 commits to 'main'
    Given the bundle server has been initialized with the remote repo
    Given I cloned from the remote repo with a bundle URI
    Given another user removed 2 commits and added 4 commits to 'main'
    Given the bundle server was updated for the remote repo
    When I fetch from the remote
    Then I am up-to-date with 'main'
    Then my repo's bundles are up-to-date with 'main'
