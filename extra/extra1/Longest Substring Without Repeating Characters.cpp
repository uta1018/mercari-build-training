#include <bits/stdc++.h>
using namespace std;
#define rep(i, n) for (int i = 0; i < n; ++i)

class Solution1 {
public:
    int lengthOfLongestSubstring(string s) {
        int ans=0;
        for(int i=0;i<s.size();i++) {
            map<char, int> Map;
            int count=0;
            for(int j=i;j<s.size();j++) {
                if(Map[s.at(j)] == 1) break;
                else {
                    Map[s.at(j)] = 1;
                    count++;
                }
            }
            ans = max(ans, count);
        }  
        return ans;
    }
};

class Solution2 {
public:
    int lengthOfLongestSubstring(string s) {
        unordered_set<char> charset;
        int left=0, right=0, ans=0;
        while(right<s.size()) {
            if(charset.contains(s[right])) {
                while(charset.contains(s[right])) {
                    charset.erase(s[left]);
                    left++;
                }
                charset.insert(s[right]);
                right++;
            } else {
                charset.insert(s[right]);
                ans = max(ans, right-left+1);
                right++;
            }
        }
        return ans;
    }
};
