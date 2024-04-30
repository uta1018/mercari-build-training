#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    vector<int> findDisappearedNumbers(vector<int>& nums) {
        for(int i=0;i<nums.size();i++) {
            int idx = abs(nums[i])-1;
            // 配列に存在する数字をインデックスとする要素を負にする
            nums[idx] = -abs(nums[idx]);
        }
        vector<int> ans;
        for(int i=0;i<nums.size();i++) {
            // 要素が負のインデックスを答えの配列に入れる
            if(nums[i] > 0) ans.push_back(i+1);
        }
        return ans;
    }
};
