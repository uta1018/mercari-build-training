#include <bits/stdc++.h>
using namespace std;
#define rep(i, n) for (int i = 0; i < n; ++i)

class Solution {
public:
    // 答えで二分探索
    int minEatingSpeed(vector<int>& piles, int h) {
        int left = 1, right = *max_element(piles.begin(), piles.end());;
        while(left < right) {
            int mid = (left+right)/2; 
            // cout << left << " " << mid << " " << right << endl;
            int hours=0;
            // 各バナナの山で、食べ切るまでの時間（切り上げ割り算）を足していく
            for(int i=0;i<piles.size();i++) {
                hours += (piles[i]+mid-1)/mid;
            }
            if(hours <= h) right = mid;
            else left = mid+1;
        }
        return left;
    }
};
