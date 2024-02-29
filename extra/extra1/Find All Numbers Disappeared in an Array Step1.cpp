#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    vector<int> findDisappearedNumbers(vector<int>& nums) {
        int n = nums.size();
        vector<int> ans;
        bool f=false;
        for(int i=1;i<=n;i++) {
            f=false;
            for(int j=0;j<n;j++) {
                if(nums[j] == i) f=true;
            }
            if(f == false) ans.push_back(i);
        }
        return ans;
    }
};
