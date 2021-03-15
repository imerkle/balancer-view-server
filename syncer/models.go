package syncer

import "github.com/hasura/go-graphql-client"

type Swap struct {
	ID             graphql.String `graphql:"id"`
	TokenIn        graphql.String `graphql:"tokenIn"`
	TokenAmountIn  graphql.String `graphql:"tokenAmountIn"`
	TokenInSym     graphql.String `graphql:"tokenInSym"`
	TokenOut       graphql.String `graphql:"tokenOut"`
	TokenOutSym    graphql.String `graphql:"tokenOutSym"`
	TokenAmountOut graphql.String `graphql:"tokenAmountOut"`
	Timestamp      graphql.Int    `graphql:"timestamp"`
}
type Pool struct {
	ID graphql.String `graphql:"id"`
}
