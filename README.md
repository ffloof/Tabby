# Tabby




# Evaluation

Eval centered around
1. Material
    - static count
2. Activity
    - weighted mobility
         - penalize squares with enemy pawn guard
         - attacks
         - should we flip this based on king sides, perhaps create several distinct maps?
3. King Safety (covered in other areas)
    - tbh king safety is not that big a deal for engines, but it leads to more fun attacking games
    - pawn shield
4. Endgames
    - weak pawns
    - passed pawns
        - opposite king passer bonus
    - eval tapered and scaled

Still need to figure out how to reward the queen moreso than other pieces for attacking king ring

# Pruning

Our current static eval is our best prediction for future static eval so it follows that...
Generally the best candidates for pruning are positions whos static eval is far from the (alpha,beta) window in either direction
These positions can be pruned usually because they are too bad/good for one side and will fall well outside the (alpha,beta) range anyway eventually, and thus are unlikely to ever become the pv
Its a waste of time to fully search these, and refuting them efficiently will greatly improve play in tactical positions with large eval changes

Most nodes in PVS are searched with a zero window and thus will either fail low or fail high, and likewise we can seperate prunings into two categories
Fail high prunings:
    - Fairly easy to predict reliably from static eval, due to null move observation
    - Less expensive especially with good move ordering
    - Prune via return/guard statements
Fail low prunings:
    - Harder to predict since tactical shots can greatly raise eval
    - Very expensive since each move will be searched in the move loop hence the fail low
    - Prune via breaking the move loop early

Most prunings can only be applied near the leaves at low depths, or else some deep idea may be overlooked by optimistic pruning. At lower depths near the leaves this idea wouldnt be discovered anyway so the pruning is justified.
 Null move pruning in many ways is the holy grail of computer chess discoveries as it is a pruning that can be applied to a position of any depth while still being reliable and pruning frequently.

# Reductions (and Extensions)

Reductions are a supplement to pruning, often some moves/positions are too risky to be outright pruned, or are close enough to the (alpha,beta) window that they could possibly become pv but it is unlikely compared to other moves.
Reductions and extensions guide the search towards exploring the more promising moves, without wasting too much time on unlikely candidates.
Pruning saves much more time as it can remove the exploration entirely, while reductions are generally much safer and more conservative though not without their own risks. 