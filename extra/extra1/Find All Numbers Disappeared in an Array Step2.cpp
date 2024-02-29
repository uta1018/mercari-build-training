#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    vector<int> findDisappearedNumbers(vector<int>& nums) {
        int n = nums.size();
        vector<int> ans;
        map<int, bool> Map;
        for(int i=0;i<n;i++) Map[nums[i]] = true;
        for(int i=1;i<=n;i++) {
            if(Map[i] != true) ans.push_back(i);
        }

        return ans;
    }
};
