#include <bits/stdc++.h>
using namespace std;
#define rep(i, n) for (int i = 0; i < n; ++i)

class Solution {
public:
    // 区間スケジューリング問題　貪欲法を使って解く
    int eraseOverlapIntervals(vector<vector<int>>& intervals) {
        // 終了時間が早い順に並び替え
        sort(intervals.begin(),intervals.end(),[](const vector<int> &alpha,const vector<int> &beta){return alpha[1] < beta[1];});
        int current=-50000, ans=0;
        for(int i=0;i<intervals.size();i++) {
            if(current > intervals[i][0]) ans++;
            else current=intervals[i][1];
        }
        return ans;
    }
};
