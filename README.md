# Tabby

This is one of my current pet projects its goal is to be a strong chess engine in as few lines of simple clean code as possible. Elegance, no code golfing, no space wasted. Its already quite strong ~2550 elo for only 800ish lines of code.

It should be compatible with any standard UCI chess interface.

It has beaten a strong National Master at my local chess club with help from a robotic arm and some computer vision that I implmented to help it play over the board.

# Game



# Chess Engine Internals 

## Evaluation

todo: cut cut cut
The idea behind static evaluation is that chess is a game far too complex to search all the way to the end of the game tree. So we have to have a heuristic way that evaluates how likely a position is to be won/drawn/lost for a given side. Its job is to understand positional patterns that will determine the game well beyond the confines of the chess engines search horizon. 

The simplest and most common way to do essentially logistic regression centered around a draw, to each position where values near 0 indicate draws and when they shift towards either infinity or -infinity indicate a position that is won/lost with increasing probability.

Most of the top engines nowadays use neural networks for statically evaluating positions however for this project I am adopting the more traditional hand crafted approach using a combination of my own chess knowledge for selecting important features and some gradient descent tuning to find the best values which can be found in `tune.py`. This is considerably more simple than the neural network approach and has the added benefit of being understandable unlike a black box, you can compute by hand the evaluation for any given position to see how the engine got its score and what features it thinks are most important. This comes at the cost of not being quite as good as a deep neural network for complex positions but even this simple approximation produces remarkably strong playing strength.

In my humble opinion there are four defining features of a position that any evaluation function has to consider, material, piece activity, king safety, and the endgame.

Material balance, is what it says on the tin do you have more pieces than your opponent, usually if a side is even ahead by a pawn this can prove to be a winning advantage.

```"It's just you and your opponent at the board and you're trying to prove something" - Bobby Fischer```

Piece activity, the quality of the pieces themselves; are they playing an active role and pressuring the opponent or are they standing around on the sidelines? This is argueably what decides most games at a high level as neither player will simply give up material unless they are under substantial accumulated pressure.

```"Help your pieces so they can help you." - Paul Morphy```

King safety, this is admittedly not as big a deal for chess engines as it is for humans since they defend so well but it leads to fun attacking games.

```"I used to attack because it was the only thing I knew. Now I attack because I know it works best." - Garry Kasparov```

Endgames, knowing which endgames are draws and which are wins saves many a game, and many features like the bishop pair or weak pawns become much more of an advantage in the endgame, while others like king safety become almost irrelevant. Understanding when its favorable to simplify into an endgame and when to avoid it is key.

```"In order to improve your game, you must study the endgame before anything else; for whereas the endings can be studied and mastered by themselves, the middlegame and the opening must be studied in relation to the endgame" - Jose Raul Capablanca```

Current evaluation outline:
1. Material
    - static count
2. Activity
    - weighted mobility
         - penalize squares with enemy pawn guard
         - bonus for attacking opponents pieces
3. King Safety (covered partially in piece activity)
    - pawn shield
4. Endgames
    - weak pawns
    - passed pawns
        - opposite king passer bonus
    - eval tapered and scaled

Still need to figure out how to reward the queen moreso than other pieces for attacking king ring

## Search

todo: write basic explanation of minmax and qsearch

## Pruning

The idea behind most prunings is we want to avoid searching positions that won't matter to the outcome of the search. The crux of all prunings is alpha beta pruning which is a branch and bound algorithm. Its a form of safe pruning designed to prune branches from the game tree that couldn't influence the outcome at the root node where we will decide on our next move to play. 

To put it simply alpha and beta represent the acceptable score bounds for each player. The idea being that if a search is can be shown to be outside of these bounds it will not be played as one of the players will make a different move further up the tree.

After we search a move the value of the position is now known to be greater than or equal to the search's score, since we know we can always make this move to achieve a certain score, and we adjust our lowerbound accordingly (opponents upperbound). 

For our subsequent moves when examining opponents responses if we find a response that exceeds our lowerbound (now their upperbound), we know that the move we made is at least worse than our original candidate move that set the bound, so we can stop examining this branch since we will end up playing our original candidate move that set the bound instead. Classic minimax by contrast would exhaustively search all lines of this move to find the true value, whereas alphabeta is content just discarding the move knowing its at least worse than a move we already considered, and thus won't be played.

In practice since we only care about the score at the root this can eliminate huge swaths of the search tree especially with good move ordering since if we consider the best moves for us and our opponent first then the bounds narrow much quicker and we reach cutoffs faster. However this is often not enough to produce a strong chess engine yet on its own (at least not one that will challenge more serious players). So we have to add some unsafe prunings to further cut the branching factor of the search.

Many positions may be absurd or fragile however it takes many moves and valuable time to find the correct refutations. These positions are very likely to fall outside the (alpha, beta) window for one side or the other, so it makes little sense to spend lots of time searching them if they are going to be pruned by alphabeta eventually. But how do we decide which positions should be pruned?

In traditional alpha beta static evaluation is only done at the leaves of the search tree and then backpropogated, however we can also perform it at the interior nodes of the tree to get a good estimate of how good/bad the position is. Generally the best candidates for pruning are positions whos static evaluation is far from the (alpha,beta) window, as the resulting search will likely also end up outside the window in either direction and this position is likely already too bad/good for one of the players and will be pruned anyway after some exploration by alphabeta. We can potentially save the time wasted on this exploration by unsafely pruning the node outright and not bothering to search it.

This is where the unsafe aspect of these prunings comes in, there are no guarantees that the position will stay the way it is, the disadvantaged side might still manage to find a brilliant resource to recover the position when we search it with high enough depth, however it is very unlikely. If we can cut enough of these lines and are right most of the time we can search deeper overall and usually find the cases where we are wrong as well. Often much of pruning is done near the leaves where even if such a brilliant resource did exist the remaining depth of the search would not be enough to find it and the result of searching and pruning is no different.

In practice these unsafe prunings work when applied under the right conditions, decrease the branching factor and greatly improve play in important tactical positions with large evaluation changes.

From the perspective of the side to move, we can broadly seperate prunings into two categories based on which bound we are pruning relative to...

Static Eval > Beta, Fail high prunings:
- Our position is already really good and we have the turn to move so its very likely we will be able to hold our advantage and produce a cutoff.
- Fairly easy to predict reliably from static eval, since being the side to move gives us the chance to defend against most existing threats (null move observation).
- Less expensive node type especially with good move ordering since most moves will fail high early.
- Prune via return/guard statements.

Static Eval < Alpha, Fail low prunings:
- Harder to predict since any move like a capture or tactical shot could suddenly greatly raise eval back to acceptable level.
- Very expensive node type since by default each move will have to be searched in the move loop to prove no such brilliant move exists.
- Usually too dangerous to prune outright instead less promissing moves are selectively skipped in the move search loop.

Generally in chess and many other similar games, while pruning fail low positions seems the most theoretically appealing its only practical near the leaves as captures and similar tactical moves can greatly improve evaluation so you do need to examine these types of moves even if the position seems at a glance lost, and as a result most pruning is concentrated on fail high positions as they are much easier to predict since you have the side to move to weasel out of any complications. 

todo: describe the common prunings used by the engine
Null move pruning in many ways is the holy grail of computer chess discoveries as it is a pruning that can be applied to a position of any depth while still being reliable and pruning frequently.

## Reductions (and Extensions)

Reductions are a supplement to pruning, often some moves/positions are too risky to be outright pruned (most fail low positions), or are close enough to the (alpha,beta) window that they could possibly become pv but it is unlikely compared to other available moves. In these cases we can reduce how deep we search certain moves based on how well we expect them to do. Reductions and extensions guide the search towards exploring the more promising moves deeply, without wasting too much time on unlikely candidates. Prunings save much more time as it can remove the exploration entirely, while reductions are generally much safer and more conservative though not without their own risks.

todo: describe the reductions and extensions the engine does