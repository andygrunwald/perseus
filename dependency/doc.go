/*
Package dependency a way to get all dependencies of a single package or application.

Dependencies are resolved with the help of so called "Resolver".
Resolver are responsible to resolve the dependencies.
They know where to find all dependencies and what are the next steps are.

Resolvers need information about single packages like the name or the url.
For this the repository.Client are there.
The repository.Client are the clients to connect to the end service and
to retrieve the necessary information.
Typically a repository.Client is a API client for services like Packagist (for PHP) or PyPI (Python).

Resolvers stream their results of the resolution via a channel back to the caller.

When we talk about dependencies we normally talking about a tree.
As an example the PHP package symfony/console:

	symfony/console
	|- symfony/polyfill-mbstring
	|- symfony/debug
		|- psr/log

Every package has n sub dependencies.
The result stream will not return a tree. It will return a list
of packages. And this algorithm is not stable.
Means: The order of the package can be different each time.
The caller code must be able to handle it.

Checkout the examples on how to use a resolver in combination with a repository.Client.
*/
package dependency
